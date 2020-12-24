package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/AllenDang/giu"
)

type LiveFile struct {
	pathname     string
	refs         []*AudioFile
	externalrefs []string
	missing      []string
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

func (scanResult *ScanResult) walk() error {

	if scanResult.rootPath == "" {
		return errors.New("rootPath is not set")
	}

	err := filepath.Walk(scanResult.rootPath,
		func(path string, info os.FileInfo, err error) error {
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

func (scanResult *ScanResult) scan(channel chan string) error {
	var err error
	var content string

	for _, file := range scanResult.liveFiles {
		channel <- fmt.Sprintf("Scanning file %s", file.pathname)
		content, err = file.gUnZipFile()
		if err != nil {
			channel <- fmt.Sprintf("File skipped - Error %s", err)
		} else {
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
	appCtx.libRootPath = "/Volumes/Work1/Musique/Sound banks/Packs Live"
	appCtx.messages = make([]string, 0)
	appCtx.scanResult = newScanResult(appCtx.scanRootPath)

	if err := appCtx.scanResult.walk(); err != nil {
		panic(err)
	}

	ch := make(chan string, 100)
	go func() {
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
