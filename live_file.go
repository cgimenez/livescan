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

//
// Unzip the file and returns content as a string
//
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

func getXMLPathElements(fileRef *xmlquery.Node, xpath string) []string {
	var dirs []string
	for _, relPathElement := range xmlquery.Find(fileRef, xpath) {
		d := relPathElement.SelectAttr("Dir")
		dirs = append(dirs, d)
	}
	return dirs
}

func dirsToPath(dirs []string) string {
	res := ""
	for _, d := range dirs {
		if d == "" {
			res += ".."
		} else {
			res += d
		}
		res += "/"
	}
	return res
}

func (livefile *LiveFile) analyzeFileRefs(scanResult *ScanResult, content string) error {
	handledSamples := make(map[string]bool)
	localSamples := make(map[string]bool)

	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return err
	}

	err = filepath.Walk(filepath.Dir(livefile.pathname),
		func(path string, info os.FileInfo, err error) error {
			switch filepath.Ext(path) {
			case ".aif", ".aiff", ".wav", ".mp3", ".mov", ".mp4":
				localSamples[filepath.Base(path)] = true
			}
			return nil
		})

	fileRefs := xmlquery.Find(doc, "//SampleRef//FileRef")
	for _, fileRef := range fileRefs {
		sampleFilename := xmlquery.Find(fileRef, "//Name")[0].SelectAttr("Value")
		hasRelativePath := xmlquery.Find(fileRef, "//HasRelativePath")[0].SelectAttr("Value") == "true"
		relativePathType := xmlquery.Find(fileRef, "//RelativePathType")[0].SelectAttr("Value")
		livePackNameElem := fileRef.SelectElement("LivePackName")

		var dirs []string
		if hasRelativePath {
			dirs = getXMLPathElements(fileRef, "//RelativePath//RelativePathElement")
		} else {
			dirs = getXMLPathElements(fileRef, "//SearchHint//PathHint//RelativePathElement")
		}

		livePackName := ""
		if livePackNameElem != nil {
			livePackName = livePackNameElem.SelectAttr("Value")
		}

		fullpath, _ := filepath.Abs(filepath.Dir(livefile.pathname) + "/" + dirsToPath(dirs) + sampleFilename)

		if handledSamples[fullpath] {
			continue
		}

		switch relativePathType {
		case "0":
		case "2":
		case "1", "3":
			if relativePathType == "1" {
				fmt.Println(dirs)
				fmt.Println(sampleFilename, fullpath)
				fmt.Println(getXMLPathElements(fileRef, "//SearchHint//PathHint//RelativePathElement"))
				fmt.Println("***")
			}
			foundLocal := localSamples[sampleFilename]
			found := scanResult.audioFiles[fullpath]
			if appCtx.fileExistsFn(fullpath) {
				audiofile, ok := scanResult.audioFiles[fullpath]
				//fmt.Println(audiofile.pathname, ok)
				if ok {
					audiofile.refs = append(audiofile.refs, livefile)
					livefile.refs = append(livefile.refs, audiofile)
				} else {
					livefile.outsiderefs = append(livefile.outsiderefs, fullpath)
				}
			} else {
				livefile.missing[fullpath] = true
			}
		case "5", "6":
			UNUSED(livePackName)
			//isSampleInPacks(sampleFilename, livePackName, dirs)
		}

		handledSamples[fullpath] = true
	}

	return nil
}
