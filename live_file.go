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

func (lifefile *LiveFile) gUnZipFile() (string, error) {
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
		switch relativePathType {
		case "5":
			dir = appCtx.libRootPath
		case "6":
			dir = filepath.Join(appCtx.libRootPath, "User Library")
		default:
			dir = filepath.Dir(lifefile_pathname)
		}
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
		dirs = append(dirs, d)
	}
	return dirs
}

func (livefile *LiveFile) analyzeFileRefs(scanResult *ScanResult, content string) error {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	for _, fileRef := range xmlquery.Find(doc, "//SampleRef//FileRef") {
		sampleFilename := xmlquery.Find(fileRef, "//Name")[0].SelectAttr("Value")
		hasRelativePath := xmlquery.Find(fileRef, "//HasRelativePath")[0].SelectAttr("Value") == "true"
		relativePathType := xmlquery.Find(fileRef, "//RelativePathType")[0].SelectAttr("Value")

		var dirs []string
		if hasRelativePath {
			dirs = getPathElements(fileRef, "//RelativePath//RelativePathElement")
		} else {
			dirs = getPathElements(fileRef, "//SearchHint//PathHint//RelativePathElement")
		}

		sampleFullPath := buildFileRefDir(livefile.pathname, sampleFilename, hasRelativePath, relativePathType, dirs)
		audiofile, ok := scanResult.audioFiles[sampleFullPath]
		//fmt.Println("dir is", ok, audiofile_path)
		if ok {
			audiofile.refs = append(audiofile.refs, livefile)
			livefile.refs = append(livefile.refs, audiofile)
		} else {
			if !fileExists(sampleFullPath) {
				fmt.Printf("Missing %s added to %s\n", sampleFullPath, livefile.pathname)
				fmt.Println(dirs, hasRelativePath, relativePathType)
				livefile.missing = append(livefile.missing, sampleFullPath)
			} else { // In packs
				livefile.librefs = append(livefile.librefs, sampleFullPath)
			}
		}
	}
	return nil
}
