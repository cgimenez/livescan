package main

import (
	"compress/gzip"
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

func buildPath(base string, directories []string, sample_filename string, hasRelativePath bool) string {
	var dir string = base

	if !hasRelativePath {
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

func findInProject(base string, directories []string, sample_filename string) (string, bool) {
	var concat_dirs string

	for i := range directories {
		concat_dirs = directories[len(directories)-1-i] + string(os.PathSeparator) + concat_dirs
		abspath, _ := filepath.Abs(filepath.Join(base, concat_dirs, sample_filename))
		if fileExists(abspath) {
			return abspath, true
		}
	}
	return "", false
}

func findOutsideProject(base string, directories []string, sample_filename string, hasRelativePath bool, relativePathType string) (string, bool) {
	var concat_dirs string

	if hasRelativePath {
		switch relativePathType {
		case "5":
			concat_dirs = appCtx.libRootPath
		case "6":
			concat_dirs = filepath.Join(appCtx.libRootPath, "User Library")
		}
	} else {
		concat_dirs = string(os.PathSeparator)
	}

	for _, d := range directories {
		if d == "" {
			d = ".."
		}
		concat_dirs = filepath.Join(concat_dirs, d)
		abspath, _ := filepath.Abs(filepath.Join(concat_dirs, sample_filename))
		if fileExists(abspath) {
			return abspath, true
		}
	}
	return "", false
}

func buildSamplePath(base string, directories []string, sample_filename string) string {
	var dir string = base

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
	var abspath string
	var foundInProject, foundOutsideProject bool

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

		var handleFound = func() {
			audiofile, ok := scanResult.audioFiles[abspath]
			if ok {
				audiofile.refs = append(audiofile.refs, livefile)
				livefile.refs = append(livefile.refs, audiofile)
			} else {
				livefile.externalrefs = append(livefile.externalrefs, abspath)
			}
		}

		abspath, foundInProject = findInProject(filepath.Dir(livefile.pathname), dirs, sampleFilename)
		if foundInProject {
			handleFound()
		}

		abspath, foundOutsideProject = findOutsideProject(filepath.Dir(livefile.pathname), dirs, sampleFilename, hasRelativePath, relativePathType)
		if foundOutsideProject {
			handleFound()
		}

		if !foundInProject && !foundOutsideProject {
			livefile.missing = append(livefile.missing, buildSamplePath(filepath.Dir(livefile.pathname), dirs, sampleFilename))
		}
	}
	return nil
}
