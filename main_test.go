package main

import (
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

/*
	live-9-p1-with-external-refs.als  references : s1.wav an external sample outside of scanned dir AND s1.wav from live-9-p2 directory
	live-9-p2-wo-external-refs.als    references : imported s1.wav, recorded 0001 1-Audio.aif
	live-10-p1-with-external-refs.als references : s1.wav an external sample outside of scanned dir
	live-10-p2-wo-external-refs.als   references : imported s1.wav, recorded 1-s1 0001 [2020-12-21 132329].aif
*/

const l9_p1 = "testdata/live-9-p1 Project/live-9-p1-with-external-refs.als"
const l9_p2 = "testdata/live-9-p2 Project/live-9-p2-wo-external-refs.als"
const l10_p1 = "testdata/live-10-p1 Project/live-10-p1-with-external-refs.als"
const l10_p2 = "testdata/live-10-p2 Project/live-10-p2-wo-external-refs.als"

const an_audio_file = "testdata/an_audio_file.wav"
const l9_p1_orphan = "/Users/chris/Desktop/s1.wav"
const l9_p2_orphan = "testdata/live-9-p2 Project/Samples/Imported/orphan.wav"
const l9_p2_s1 = "testdata/live-9-p2 Project/Samples/Imported/s1.wav"
const l9_p2_recorded = "testdata/live-9-p2 Project/Samples/Recorded/0001 1-Audio.aif"
const l10_p1_orphan = "/Users/chris/Desktop/s1.wav"
const l10_p2_orphan = "testdata/live-10-p2 Project/Samples/Imported/orphan.wav"
const l10_p2_s1 = "testdata/live-10-p2 Project/Samples/Imported/s1.wav"
const l10_p2_recorded = "testdata/live-10-p2 Project/Samples/Recorded/1-s1 0001 [2020-12-21 132329].aif"

var audioFileNames = [...]string{
	an_audio_file,
	l9_p1_orphan,
	l10_p2_orphan,
	l10_p2_s1,
	l10_p2_recorded,
	l9_p2_orphan,
	l9_p2_s1,
	l9_p2_recorded,
}

var projectFileNames = [...]string{l9_p1, l9_p2, l10_p1, l10_p2}

func array_contains(arr []string, b interface{}) bool {
	for _, a := range arr {
		if a == b {
			return true
		}
	}
	return false
}

func callingLine() int {
	_, _, no, _ := runtime.Caller(2)
	return no
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
		//panic("should not be nil")
	}
}

func assert_equal(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("%v not equal to %v - line %d", a, b, callingLine())
	}
}

func assertAudioFileOwned(t *testing.T, a *audioFile, l *liveFile) {
	var found bool = false

	for _, ref := range l.refs {
		if a == ref {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected audiofile %s to be owned by livefile %s - line %d", a.pathname, l.pathname, callingLine())
	}
}

var scr *scanResult = nil

func setup(t *testing.T) {

	err := os.RemoveAll("testdata/live-10-p1 Project/Backup/")
	if err != nil {
		log.Fatal(err)
	}

	err = os.RemoveAll("testdata/live-10-p2 Project/Backup/")
	if err != nil {
		log.Fatal(err)
	}

	for i := range projectFileNames {
		fabs, _ := filepath.Abs(projectFileNames[i])
		projectFileNames[i] = fabs
	}

	for i := range audioFileNames {
		if !filepath.IsAbs(audioFileNames[i]) {
			fabs, _ := filepath.Abs(audioFileNames[i])
			audioFileNames[i] = fabs
		}
	}

	scr = newScanResult("./testdata")
}

func teardown() {
}

func TestScanRoot(t *testing.T) {
	t.Run("Scanned files must be found", func(t *testing.T) {
		setup(t)

		err := scr.walk()
		if err != nil {
			t.Error(err)
		}
		err = scr.scan(make(chan string, 100))
		if err != nil {
			t.Error(err)
		}

		for k := range scr.audioFiles {
			found := array_contains(audioFileNames[:], k)
			if !found {
				t.Errorf("Expected %s to be in found in audioFileNames", k)
			}
		}

		for k := range scr.liveFiles {
			found := array_contains(projectFileNames[:], k)
			if !found {
				t.Errorf("Expected %s to be in found in projectFileNames", k)
			}
		}

		teardown()
	})

}

func TestFileRefs(t *testing.T) {
	setup(t)

	err := scr.walk()
	if err != nil {
		t.Error(err)
	}

	err = scr.scan(make(chan string, 100))
	if err != nil {
		t.Error(err)
	}

	for _, projectFileName := range projectFileNames {
		livefile := scr.liveFiles[projectFileName]

		switch projectFileName {
		case l9_p1:
			assert_equal(t, 2, len(livefile.orphans))
			assert_array_contains(t, livefile.orphans, l9_p1_orphan)
			assertAudioFileOwned(t, scr.audioFiles[l9_p2_s1], livefile)
		case l9_p2:
			assertAudioFileOwned(t, scr.audioFiles[l9_p2_s1], livefile)
			assertAudioFileOwned(t, scr.audioFiles[l9_p2_recorded], livefile)
		case l10_p1:
			assert_equal(t, 0, len(livefile.refs)) // No samples
			assert_array_contains(t, livefile.orphans, l10_p1_orphan)
		case l10_p2:
			assertAudioFileOwned(t, scr.audioFiles[l10_p2_s1], livefile)
			assertAudioFileOwned(t, scr.audioFiles[l10_p2_recorded], livefile)
		}
	}
	teardown()
}

func TestBuildFileRefDir(t *testing.T) {
	dirs := []string{"foo", "bar", "joe"}
	dir := buildFileRefDir("/usr/local/projects/project1/project1.als", "sample1.wav", false, "", dirs)
	assert_equal(t, "/foo/bar/joe/sample1.wav", dir)

	dir = buildFileRefDir("/usr/local/projects/project1/project1.als", "sample1.wav", true, "", dirs)
	assert_equal(t, "/usr/local/projects/project1/foo/bar/joe/sample1.wav", dir)
}

func TestAll(t *testing.T) {
	setup(t)

	/* for _, audioFile := range scr.audioFiles {
		fmt.Println(audioFile.pathname)
		if len(audioFile.refs) == 0 {
			fmt.Println(" No refs")
		} else {
			for _, ref := range audioFile.refs {
				fmt.Printf(" %s\n", (*ref).pathname)
			}
		}
	}
	println("----------------------------------------------")
	for _, liveFile := range scr.liveFiles {
		fmt.Println(liveFile.pathname)
		if len(liveFile.refs) == 0 {
			fmt.Println(" No refs")
		} else {
			for _, ref := range liveFile.refs {
				fmt.Printf(" %s\n", (*ref).pathname)
			}
		}
	}

	if len(scr.audioFiles[an_audio_file].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	}

	if len(scr.audioFiles[l9_p2_orphan].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	}

	if len(scr.audioFiles[l10_p2_orphan].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	} */

	teardown()
}
