package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/AllenDang/giu"
	"github.com/antchfx/xmlquery"
)

type LiveFile struct {
	pathname    string
	refs        []*AudioFile
	outsiderefs []string
	missing     []string
}

type AudioFile struct {
	pathname string
	size     int64
	refs     []*LiveFile
}

type AudioFileMap map[string]*AudioFile
type LiveFileMap map[string]*LiveFile

type ScanResult struct {
	rootPath       string
	liveFiles      LiveFileMap
	liveFilesKeys  []string
	audioFiles     AudioFileMap
	audioFilesKeys []string
	audioFilesSize int64
}

func (scanResult *ScanResult) sort() {
	scanResult.audioFilesKeys = make([]string, len(scanResult.audioFiles))
	i := 0
	for k := range scanResult.audioFiles {
		scanResult.audioFilesKeys[i] = k
		i++
	}
	sort.Strings(scanResult.audioFilesKeys)

	scanResult.liveFilesKeys = make([]string, len(scanResult.liveFiles))
	i = 0
	for k := range scanResult.liveFiles {
		scanResult.liveFilesKeys[i] = k
		i++
	}
	sort.Strings(scanResult.liveFilesKeys)
}

func newScanResult(rootPath string) *ScanResult {
	sc := new(ScanResult)
	sc.rootPath = rootPath
	sc.liveFiles = make(LiveFileMap)
	sc.audioFiles = make(AudioFileMap)
	return sc
}

//
// Scan the directory provided by the user to find all .als and audio files
//
func (scanResult *ScanResult) walk() error {

	if scanResult.rootPath == "" {
		return errors.New("rootPath is not set")
	}

	err := filepath.Walk(scanResult.rootPath,
		func(path string, info os.FileInfo, err error) error {
			match, _ := regexp.MatchString(`.*\[\d+-\d+-\d+\s\d+\]\.als`, filepath.Base(path))
			if match { // Skip live 10 backups
				return filepath.SkipDir
			}
			path, _ = filepath.Abs(path)
			base := filepath.Base(path)
			if err != nil && !isIconFile(base) { // Icon? files cause problems under Windows
				return fmt.Errorf("name %s: %v", base, err)
			}
			switch filepath.Ext(path) {
			case ".aif", ".wav", ".mp3":
				scanResult.audioFiles[path] = &AudioFile{path, info.Size(), make([]*LiveFile, 0)}
			case ".als":
				scanResult.liveFiles[path] = &LiveFile{path, make([]*AudioFile, 0), make([]string, 0), make([]string, 0)}
			}
			return nil
		})
	return err
}

//
// gunzip and analyse the .als files found during walk() process
//
func (scanResult *ScanResult) scan(channel chan string) error {
	var err error
	var content string

	for _, file := range scanResult.liveFiles {
		content, err = file.gUnZipFile()
		if err != nil {
			channel <- fmt.Sprintf("File skipped %s - Error %s", file.pathname, err)
		} else {
			channel <- fmt.Sprintf("Scanning file %s", file.pathname)
			err = file.analyzeFileRefs(scanResult, content)
			if err != nil {
				close(channel)
				return err
			}
		}
	}
	close(channel)
	scanResult.sort()
	return nil
}

func startScan(uiMode int) {
	appCtx.messages = make([]string, 0)
	appCtx.messages = append(appCtx.messages, "Finding files...")
	appCtx.scanResult = newScanResult(appCtx.scanRootPath)

	ch := make(chan string, 100)
	go func() {
		if err := appCtx.scanResult.walk(); err != nil {
			panic(err)
		}
		if err := appCtx.scanResult.scan(ch); err != nil {
			panic(err)
		}
	}()

	appCtx.scanStatus = RUNNING
	for appCtx.scanStatus == RUNNING {
		msg, more := <-ch
		if more {
			if uiMode == GUI {
				appCtx.messages = append(appCtx.messages, msg)
			} else {
				log.Println(msg)
			}
		} else {
			appCtx.scanStatus = DONE
		}
		if uiMode == GUI {
			giu.Update()
		}
	}
}

func (scanResult *ScanResult) debug() {
	/*
		The logic seems to be
		relativePathType 0 && HasRelativePath false : ? Missing sample ?
		relativePathType 1 && HasRelativePath true  : sample is relative to project but outside of project directory [but can appear twice in the project, with 3 / true OR 6 / true]
		relativePathType 2 && HasRelativePath true  : ? Missing sample ?
		relativePathType 3 && HasRelativePath true  : sample is under project's Samples dir
		relativePathType 5 && HasRelativePath true  : sample is in the Ableton Live's packs directory but pathname might be broken (must search in packs)
		relativePathType 6 && HasRelativePath true  : sample is in User Pack directory but sometimes is under project's Samples dir

		In any case, if the audio file is present under Samples project dir, it must have precedence
	*/
	var content string
	var err error

	if err = appCtx.scanResult.walk(); err != nil {
		panic(err)
	}
	for _, file := range scanResult.liveFiles {
		content, err = file.gUnZipFile()
		if err != nil {
			fmt.Println(err)
		} else {
			doc, err := xmlquery.Parse(strings.NewReader(content))
			if err != nil {
				fmt.Println(err)
			}

			for _, fileRef := range xmlquery.Find(doc, "//SampleRef//FileRef") {
				sampleFilename := xmlquery.Find(fileRef, "//Name")[0].SelectAttr("Value")
				hasRelativePath := xmlquery.Find(fileRef, "//HasRelativePath")[0].SelectAttr("Value") == "true"
				relativePathType := xmlquery.Find(fileRef, "//RelativePathType")[0].SelectAttr("Value")
				livePackName := xmlquery.Find(fileRef, "//LivePackName")[0].SelectAttr("Value")
				livePackID := xmlquery.Find(fileRef, "//LivePackId")[0].SelectAttr("Value")

				var rel_dirs, abs_dirs []string
				rel_dirs = getPathElements(fileRef, "//RelativePath//RelativePathElement")
				abs_dirs = getPathElements(fileRef, "//SearchHint//PathHint//RelativePathElement")
				rel_path, abs_path := "", ""
				for i := range rel_dirs {
					if rel_dirs[i] == "" {
						rel_path += "../"
					} else {
						rel_path += rel_dirs[i] + "/"
					}
				}
				for i := range abs_dirs {
					abs_path += abs_dirs[i] + "/"
				}
				fmt.Printf("%s\t%s\t%s\t%t\t%s\t%s\t%s\t%s\n", file.pathname, sampleFilename, relativePathType, hasRelativePath, livePackName, livePackID, rel_path, abs_path)
			}
		}
	}
}

//
//
func list() {
	if len(appCtx.scanResult.audioFiles) > 0 {
		log.Println()
		log.Println("Orphan audio files :")
		for _, audiofile := range appCtx.scanResult.audioFiles {
			log.Printf(" %s\n", audiofile.pathname)
		}
	}

	for _, livefile := range appCtx.scanResult.liveFiles {
		log.Println()
		log.Println(livefile.pathname)
		for _, ref := range livefile.refs {
			log.Println(" Ref :", ref.pathname)
		}
		for _, missing := range livefile.missing {
			log.Println(" MIS :", missing)
		}
	}
}

func (scanResult *ScanResult) deleteAudioFile(audiofilePath string) {
	var i int
	for i = range scanResult.audioFilesKeys {
		if scanResult.audioFilesKeys[i] == audiofilePath {
			break
		}
	}
	delete(scanResult.audioFiles, audiofilePath)
	scanResult.audioFilesKeys = append(scanResult.audioFilesKeys[:i], scanResult.audioFilesKeys[i+1:]...)
}
