package main

import (
	"fmt"
	"testing"
)

/*
	live-9-p1-with-external-refs.als  references : s1.wav an external sample outside of scanned dir AND s1.wav from live-9-p2-wo-external-refs.als
	live-9-p2-wo-external-refs.als    references : s1.wav, 0001 1-Audio.aif
	live-10-p1-with-external-refs.als references : an external sample outside of scanned dir
	live-10-p2-wo-external-refs.als   references : s1.wav, 1-s1 0001 [2020-12-21 132329].aif
*/

const l9_p1 = "testdata/live-9-p1 Project/live-9-p1-with-external-refs.als"
const l9_p2 = "testdata/live-9-p2 Project/live-9-p2-wo-external-refs.als"
const l10_p1 = "testdata/live-10-p1 Project/live-10-p1-with-external-refs.als"
const l10_p2 = "testdata/live-10-p2 Project/live-10-p2-wo-external-refs.als"

const an_audio_file = "testdata/an_audio_file.wav"
const l9_p2_orphan = "testdata/live-9-p2 Project/Samples/Imported/orphan.wav"
const l9_p2_s1 = "testdata/live-9-p2 Project/Samples/Imported/s1.wav"
const l9_p2_recorded = "testdata/live-9-p2 Project/Samples/Recorded/0001 1-Audio.aif"
const l10_p2_orphan = "testdata/live-10-p2 Project/Samples/Imported/orphan.wav"
const l10_p2_s1 = "testdata/live-10-p2 Project/Samples/Imported/s1.wav"
const l10_p2_recorded = "testdata/live-10-p2 Project/Samples/Recorded/1-s1 0001 [2020-12-21 132329].aif"

var audioFileNames = [...]string{
	an_audio_file,
	l10_p2_orphan,
	l10_p2_s1,
	l10_p2_recorded,
	l9_p2_orphan,
	l9_p2_s1,
	l9_p2_recorded,
}

var projectFileNames = [...]string{l9_p1, l9_p2, l10_p1, l10_p2,
	"testdata/live-10-p1 Project/Backup/live-10-p1-with-external-refs [2020-12-21 131739].als",
	"testdata/live-10-p2 Project/Backup/live-10-p2-wo-external-refs [2020-12-21 132411].als",
}

var files filesT

func array_contains(arr []string, str string) bool {
	for _, a := range arr {
		if string(a) == str {
			return true
		}
	}
	return false
}

func setup(t *testing.T) {
	var err error

	files, err = scanDirectory("./testdata")
	if err != nil {
		t.Error(err)
	}
}

func teardown() {
}

func TestScanRoot(t *testing.T) {
	t.Run("Scanned files must be found", func(t *testing.T) {
		setup(t)

		for k := range files.audioFiles {
			found := array_contains(audioFileNames[:], k)
			if !found {
				t.Errorf("Expected %s to be in found in audioFileNames", k)
			}
		}

		for k := range files.liveFiles {
			found := array_contains(projectFileNames[:], k)
			if !found {
				t.Errorf("Expected %s to be in found in projectFileNames", k)
			}
		}

		teardown()
	})

}

func TestFileRefsCount(t *testing.T) {
	setup(t)

	expected_f := []string{l9_p1, l9_p2, l10_p1, l10_p2}

	for index := 0; index < len(expected_f); index++ {
		livef := files.liveFiles[expected_f[index]]

		content, err := gUnZipFile(*livef)
		if err != nil {
			t.Error(err)
		}

		err = analyzeFileRefs(&files, livef, content)
		if err != nil {
			t.Error(err)
		}
	}
	teardown()
}

func TestAll(t *testing.T) {
	setup(t)

	err := scanLiveProjects(&files, nil)

	if err != nil {
		t.Error(err)
	}

	for _, audioFile := range files.audioFiles {
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
	for _, liveFile := range files.liveFiles {
		fmt.Println(liveFile.pathname)
		if len(liveFile.refs) == 0 {
			fmt.Println(" No refs")
		} else {
			for _, ref := range liveFile.refs {
				fmt.Printf(" %s\n", (*ref).pathname)
			}
		}
	}

	if len(files.audioFiles[an_audio_file].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	}

	if len(files.audioFiles[l9_p2_orphan].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	}

	if len(files.audioFiles[l10_p2_orphan].refs) != 0 {
		t.Errorf("%s should be orphan", an_audio_file)
	}

	teardown()
}
