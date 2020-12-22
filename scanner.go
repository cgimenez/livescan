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

func (sc *scanResult) scan(displayFn func(string)) error {
	for _, file := range sc.liveFiles {
		if displayFn != nil {
			displayFn(fmt.Sprintf("Scanning file %s", file.pathname))
		}
		content, err := file.gUnZipFile()
		if err != nil && displayFn != nil {
			displayFn(fmt.Sprintf("File skipped - Error %s", err))
		}
		err = file.analyzeFileRefs(sc, content)
		if err != nil {
			return err
		}
	}
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
	} else {
		dir = lifefile_pathname
		for _, d := range directories {
			dir += "/" + d
		}
		dir += "/" + sample_filename
	}

	return dir
}

func (livefile *liveFile) analyzeFileRefs(sc *scanResult, content string) error {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	for _, ref := range xmlquery.Find(doc, "//SampleRef//FileRef") {
		var rebuilt_pathname string

		name := xmlquery.Find(ref, "//Name")[0].SelectAttr("Value")
		//fmt.Println("New ref for", name)

		hasRelativePath := xmlquery.Find(ref, "//HasRelativePath")[0].SelectAttr("Value") == "true"
		//relativePathType := xmlquery.Find(ref, "//RelativePathType")[0].SelectAttr("Value")
		if hasRelativePath {
			rebuilt_pathname = filepath.Dir(livefile.pathname)
			for _, relPathElement := range xmlquery.Find(ref, "//RelativePath//RelativePathElement") {
				if relPathElement.SelectAttr("Dir") != "" {
					rebuilt_pathname += "/" + relPathElement.SelectAttr("Dir")
				} else {
					rebuilt_pathname += "/.."
				}
			}
		} else {
			for _, relPathElement := range xmlquery.Find(ref, "//SearchHint//PathHint//RelativePathElement") {
				rebuilt_pathname += "/" + relPathElement.SelectAttr("Dir")
			}
		}
		rebuilt_pathname += "/" + filepath.Base(name)
		audiofile, ok := sc.audioFiles[rebuilt_pathname]
		if ok {
			audiofile.refs = append(audiofile.refs, livefile)
			livefile.refs = append(livefile.refs, audiofile)
		} else {
			//fmt.Printf("add orphan %s to %s\n", rebuilt_pathname, livefile.pathname)
			livefile.orphans = append(livefile.orphans, rebuilt_pathname)
		}
	}
	return nil
}

// From https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
