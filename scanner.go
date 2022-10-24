package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/AllenDang/giu"
)

type LiveFile struct {
	pathname    string
	refs        []*AudioFile
	outsiderefs []string
	missing     map[string]bool
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
func (scanResult *ScanResult) walk(channel chan string) error {

	if scanResult.rootPath == "" {
		return errors.New("rootPath is not set")
	}

	err := filepath.Walk(scanResult.rootPath,
		func(path string, info os.FileInfo, err error) error {
			match, _ := regexp.MatchString(`.*\[\d+-\d+-\d+\s\d+\]\.als`, filepath.Base(path))
			if match { // Skip live 10 backups
				return filepath.SkipDir
			}
			if info.IsDir() {
				channel <- fmt.Sprintf("Scanning dir %s", path)
			}
			path, _ = filepath.Abs(path)
			base := filepath.Base(path)
			if err != nil && !isIconFile(base) { // Icon? files cause problems under Windows
				return fmt.Errorf("name %s: %v", base, err)
			}
			switch filepath.Ext(path) {
			case ".aif", ".aiff", ".wav", ".mp3", ".mov", ".mp4":
				scanResult.audioFiles[path] = &AudioFile{path, info.Size(), make([]*LiveFile, 0)}
			case ".als":
				scanResult.liveFiles[path] = &LiveFile{path, make([]*AudioFile, 0), make([]string, 0), make(map[string]bool)}
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
		if err := appCtx.scanResult.walk(ch); err != nil {
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
				fmt.Println(msg)
			}
		} else {
			appCtx.scanStatus = DONE
		}
		if uiMode == GUI {
			giu.Update()
		}
	}
}

//
//
func (scanResult *ScanResult) list() {
	if len(scanResult.audioFiles) > 0 {
		fmt.Println("AUDIO FILES FOUND")
		for _, audiofile := range scanResult.audioFiles {
			if len(audiofile.refs) == 0 {
				fmt.Print("ORPHAN : ")
			}
			fmt.Println(audiofile.pathname)
		}
	}

	//fmt.Println(appCtx.scanResult.audioFiles)
	for _, livefile := range appCtx.scanResult.liveFiles {
		fmt.Println()
		fmt.Println(livefile.pathname)
		for _, ref := range livefile.refs {
			fmt.Println(" REF :", ref.pathname)
		}
		for key, _ := range livefile.missing {
			fmt.Println(" MIS :", key)
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
