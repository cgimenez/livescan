package main

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/xmlquery"
)

type liveFile struct {
	pathname string
	refs     []*audioFile
	orphans  []string
}

type audioFile struct {
	pathname string
	size     int64
	refs     []*liveFile
}

type audioFileMap map[string]*audioFile
type liveFileMap map[string]*liveFile

type scanResult struct {
	rootPath   string
	liveFiles  liveFileMap
	audioFiles audioFileMap
}

func newScanResult(rootPath string) *scanResult {
	sc := new(scanResult)
	sc.rootPath = rootPath
	sc.liveFiles = make(liveFileMap)
	sc.audioFiles = make(audioFileMap)
	return sc
}

func (sc *scanResult) walk() error {

	if sc.rootPath == "" {
		return errors.New("rootPath is not set")
	}

	err := filepath.Walk(sc.rootPath,
		func(path string, info os.FileInfo, err error) error {
			path, _ = filepath.Abs(path)
			if err != nil {
				return err
			}
			switch filepath.Ext(path) {
			case ".aif", ".wav", ".mp3":
				sc.audioFiles[path] = &audioFile{path, info.Size(), make([]*liveFile, 0)}
			case ".als":
				sc.liveFiles[path] = &liveFile{path, make([]*audioFile, 0), make([]string, 0)}
			}
			return nil
		})
	return err
}

func (sc *scanResult) scan(channel chan string) error {
	var err error
	var content string

	for _, file := range sc.liveFiles {
		channel <- fmt.Sprintf("Scanning file %s", file.pathname)
		content, err = file.gUnZipFile()
		if err != nil {
			channel <- fmt.Sprintf("File skipped - Error %s", err)
		} else {
			err = file.analyzeFileRefs(sc, content)
			if err != nil {
				close(channel)
				return err
			}
		}
	}
	close(channel)
	return nil
}

func (lifefile *liveFile) gUnZipFile() (string, error) {
	f, err := os.Open(lifefile.pathname)
	if err != nil {
		return "", err
	}

	zr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}

	defer func() {
		f.Close()
		zr.Close()
	}()

	content, err := ioutil.ReadAll(zr)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func buildFileRefDir(lifefile_pathname string, sample_filename string, hasRelativePath bool, relativePathType string, directories []string) string {
	var dir string

	if hasRelativePath {
		dir = filepath.Dir(lifefile_pathname)
	} else {
		dir = "/"
	}
	for _, d := range directories {
		if d == "" {
			d = ".."
		}
		dir = filepath.Join(dir, d)
	}
	dir, _ = filepath.Abs(filepath.Join(dir, sample_filename))

	return dir
}

func getPathElements(fileRef *xmlquery.Node, xpath string) []string {
	var dirs []string
	for _, relPathElement := range xmlquery.Find(fileRef, xpath) {
		d := relPathElement.SelectAttr("Dir")
		//if len(d) > 0 {
		dirs = append(dirs, d)
		//}
	}
	return dirs
}

func (livefile *liveFile) analyzeFileRefs(sc *scanResult, content string) error {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	for _, fileRef := range xmlquery.Find(doc, "//SampleRef//FileRef") {
		sample_filename := xmlquery.Find(fileRef, "//Name")[0].SelectAttr("Value")
		hasRelativePath := xmlquery.Find(fileRef, "//HasRelativePath")[0].SelectAttr("Value") == "true"
		relativePathType := xmlquery.Find(fileRef, "//RelativePathType")[0].SelectAttr("Value")

		var dirs []string
		if hasRelativePath {
			dirs = getPathElements(fileRef, "//RelativePath//RelativePathElement")
		}
		if !hasRelativePath || len(dirs) == 0 {
			if len(dirs) == 0 {
				hasRelativePath = false
			}
			dirs = getPathElements(fileRef, "//SearchHint//PathHint//RelativePathElement")
		}

		audiofile_path := buildFileRefDir(livefile.pathname, sample_filename, hasRelativePath, relativePathType, dirs)
		if livefile.pathname == "testdata/live-9-p1 Project/live-9-p1-with-external-refs.als" {
			fmt.Println(dirs, audiofile_path)
		}
		audiofile, ok := sc.audioFiles[audiofile_path]
		//fmt.Println("dir is", ok, audiofile_path)
		if ok {
			audiofile.refs = append(audiofile.refs, livefile)
			livefile.refs = append(livefile.refs, audiofile)
		} else {
			//fmt.Printf("Orphan %s added to %s\n", audiofile_path, livefile.pathname)
			livefile.orphans = append(livefile.orphans, audiofile_path)
		}
	}
	return nil
}
