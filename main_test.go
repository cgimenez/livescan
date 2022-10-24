package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

func callingLine() int {
	_, _, no, _ := runtime.Caller(2)
	return no
}

func absPath(path string) string {
	fabs, _ := filepath.Abs(path)
	return fabs
}

func assert_array_contains(t *testing.T, arr []string, b string) {
	res := false
	for _, a := range arr {
		if a == b {
			res = true
		}
	}
	if !res {
		t.Errorf("%s not found - line %d", b, callingLine())
	}
}

func assert_not_nil(t *testing.T, v interface{}) {
	if v == nil || reflect.ValueOf(v).IsNil() {
		t.Errorf("should not be nil")
	}
}

func assert_equal(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("expected %v got %v - line %d", a, b, callingLine())
	}
}

func assertAudioFileOwned(t *testing.T, scr *ScanResult, audioFilePath string, liveFilePath string) {
	audioFilePath, _ = filepath.Abs(audioFilePath)
	liveFilePath, _ = filepath.Abs(liveFilePath)

	audioFile := scr.audioFiles[audioFilePath]
	liveFile := scr.liveFiles[liveFilePath]

	found := false

	for _, ref := range liveFile.refs {
		if audioFile == ref {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected audiofile %s to be owned by livefile %s - line %d", audioFilePath, liveFilePath, callingLine())
	}
}

func buildXML(dirs []string, sample_filename string, hasRelativePath bool, relativePathType string) string {
	template := `
	<SampleRef>
		<FileRef>
			<HasRelativePath Value="%t" />
			<RelativePathType Value="%s" />
			<RelativePath>
				%s
			</RelativePath>
			<Name Value="%s" />
			<RefersToFolder Value="false" />
			<SearchHint>
				<PathHint>
					%s
				</PathHint>
			</SearchHint>
			<LivePackName Value="" />
			<LivePackId Value="" />
	</FileRef>
</SampleRef>
`
	sdirs := ""
	for _, d := range dirs {
		sdirs += fmt.Sprintf(`<RelativePathElement Dir="%s" />`, d) + "\n"
	}

	xml := ""
	if hasRelativePath {
		xml = fmt.Sprintf(template, hasRelativePath, relativePathType, sdirs, sample_filename, "")
	} else {
		xml = fmt.Sprintf(template, hasRelativePath, relativePathType, "", sample_filename, sdirs)
	}

	return xml
}

var scr *ScanResult = nil
var liveFile *LiveFile

func setup() {
	scr = newScanResult("./testdata")
	//liveFile = &LiveFile{"test.als", make([]*AudioFile, 0), make([]string, 0), make(map[string]bool)}
}

func teardown() {
}

func TestAnalyzeFileRefs(t *testing.T) {
	t.Run("Sample inside project", func(t *testing.T) {
		setup()
		appCtx.fileExistsFn = func(filename string) bool {
			if filename == absPath("Samples/Recorded/s1.wav") {
				return true
			} else {
				return false
			}
		}
		samplePath := absPath("Samples/Recorded/s1.wav")
		scr.audioFiles[samplePath] = &AudioFile{samplePath, 0, make([]*LiveFile, 0)}
		scr.liveFiles["foo.als"] = &LiveFile{"foo.als", make([]*AudioFile, 0), make([]string, 0), make(map[string]bool)}
		xml := buildXML([]string{"Samples", "Recorded"}, "s1.wav", true, "3")
		liveFile.analyzeFileRefs(scr, xml)
		assert_equal(t, 1, len(liveFile.refs))
		assert_equal(t, liveFile.refs[0], scr.audioFiles[samplePath])
		teardown()
	})

	t.Run("Sample outside and relative to project", func(t *testing.T) {
		setup()
		appCtx.fileExistsFn = func(filename string) bool {
			if filename == absPath("../../Somewhere/Samples/Recorded/s1.wav") {
				return true
			} else {
				return false
			}
		}
		samplePath := absPath("../../Somewhere/Samples/Recorded/s1.wav")
		scr.audioFiles[samplePath] = &AudioFile{samplePath, 0, make([]*LiveFile, 0)}
		scr.liveFiles["foo.als"] = &LiveFile{"foo.als", make([]*AudioFile, 0), make([]string, 0), make(map[string]bool)}
		xml := buildXML([]string{"", "", "Somewhere", "Samples", "Recorded"}, "s1.wav", true, "3")
		liveFile.analyzeFileRefs(scr, xml)
		assert_equal(t, 1, len(liveFile.refs))
		assert_equal(t, liveFile.refs[0], scr.audioFiles[samplePath])
		teardown()
	})
}

func TestPacks(t *testing.T) {
	t.Run("Find sample in packs", func(t *testing.T) {
		var err error

		setup()
		appCtx.env = TEST
		setAppCtxDefaults()
		appCtx.libRootPath = "/Somewhere/Packs Live"
		isLivePacksDBScanned()
		if _, err = appCtx.dbCnx.Exec("DELETE FROM packs_samples"); err != nil {
			t.Error(err)
		}

		data := map[string]string{"pack1/metal/clap/samples/processed": "s1.wav", "pack2/plastic/boom/samples/recorded": "s1.wav"}
		for p, s := range data {
			if _, err = appCtx.dbCnx.Exec("INSERT INTO packs_samples (directory, name) VALUES ((?), (?))", appCtx.libRootPath+"/"+p, s); err != nil {
				t.Error(err)
			}
		}

		b, _ := isSampleInPacks("s1.wav", "pack1", []string{"metal", "clap", "samples", "processed"})
		assert_equal(t, true, b)
		b, _ = isSampleInPacks("s2.wav", "pack1", []string{"metal", "clap", "samples", "processed"})
		assert_equal(t, false, b)
		teardown()
	})
}
