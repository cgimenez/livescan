package main

import (
	"flag"
	"fmt"
	"log"
)

const rootPathName string = "/Volumes/Work1/Musique/2019-Curated-2"

func displayMsg(msg string) {
	log.Println(msg)
}

func startScan() {
	sc := newScanResult(rootPathName)
	err := sc.walk()
	if err != nil {
		panic(err)
	}

	if err = sc.scan(displayMsg); err != nil {
		panic(err)
	}

	log.Println("Orphan audio files")
	var sz int64
	for _, audioFile := range sc.audioFiles {
		if len(audioFile.refs) == 0 {
			log.Printf(audioFile.pathname)
			sz += audioFile.size
		}
	}

	fmt.Printf("Size to clean %s\n", ByteCountSI(sz))

	/*for _, liveFile := range files.liveFiles {
		if len(liveFile.refs) == 0 {
			log.Printf(liveFile.pathname)
		}
	}*/
}

func main() {
	noGUI := flag.Bool("no-gui", false, "disable GUI")
	flag.Parse()

	if *noGUI {
		startScan()
	} else {
		setupGUI()
	}

}
