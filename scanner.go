package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/antchfx/xmlquery"
)

type liveFileT struct {
	pathname string
	refs     []*audioFileT
}

type audioFileT struct {
	pathname string
	size     int64
	refs     []*liveFileT
}

type audioFileMap map[string]*audioFileT
type liveFileMap map[string]*liveFileT

type filesT struct {
	liveFiles  liveFileMap
	audioFiles audioFileMap
}

func scanDirectory(root_path string) (filesT, error) {

	files := filesT{make(liveFileMap), make(audioFileMap)}

	err := filepath.Walk(root_path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			switch filepath.Ext(path) {
			case ".aif", ".wav", ".mp3":
				files.audioFiles[path] = &audioFileT{path, info.Size(), make([]*liveFileT, 0)}
			case ".als":
				files.liveFiles[path] = &liveFileT{path, make([]*audioFileT, 0)}
			}
			return nil
		})
	return files, err
}

func gUnZipFile(file liveFileT) (string, error) {
	f, err := os.Open(file.pathname)
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

func analyzeFileRefs(files *filesT, liveFile *liveFileT, content string) error {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	for _, ref := range xmlquery.Find(doc, "//SampleRef//FileRef") {
		var rebuilt_pathname string

		name := xmlquery.Find(ref, "//Name")[0].SelectAttr("Value")
		//fmt.Println("New ref for", name)

		hasRelativePath := xmlquery.Find(ref, "//HasRelativePath")[0].SelectAttr("Value") == "true"
		if hasRelativePath {
			rebuilt_pathname = filepath.Dir(liveFile.pathname)
			for _, relPathElement := range xmlquery.Find(ref, "//RelativePath//RelativePathElement") {
				rebuilt_pathname += "/" + relPathElement.SelectAttr("Dir")
			}
		} else {
			for _, relPathElement := range xmlquery.Find(ref, "//SearchHint//PathHint//RelativePathElement") {
				rebuilt_pathname += "/" + relPathElement.SelectAttr("Dir")
			}
		}
		rebuilt_pathname += "/" + filepath.Base(name)
		//fmt.Println("rebuilt_pathname will be", rebuilt_pathname)
		audioFile, ok := files.audioFiles[rebuilt_pathname]
		if ok {
			//fmt.Println("found")
			audioFile.refs = append(audioFile.refs, liveFile)
			liveFile.refs = append(liveFile.refs, audioFile)
		}
	}
	return nil
}

func scanLiveProjects(files *filesT, displayFn func(string)) error {
	for _, file := range files.liveFiles {
		if displayFn != nil {
			displayFn(fmt.Sprintf("Scanning file %s", file.pathname))
		}
		content, err := gUnZipFile(*file)
		if err != nil && displayFn != nil {
			displayFn(fmt.Sprintf("File skipped - Error %s", err))
		}
		err = analyzeFileRefs(files, file, content)
		if err != nil {
			return err
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
