package main

import (
	"flag"
	"fmt"
	"log"
)

const rootPathName string = "/Volumes/Work1/Musique/Génériques Pour la Nouvelle Télévision/Compos/Bloc 4"

func startScan() {
	sc := newScanResult(rootPathName)

	log.Println("Finding files")
	err := sc.walk()
	if err != nil {
		panic(err)
	}

	ch := make(chan string, 100)
	go func() {
		if err = sc.scan(ch); err != nil {
			panic(err)
		}
	}()

	done := false
	for !done {
		msg, more := <-ch
		if more {
			log.Println(msg)
		} else {
			fmt.Println("Scan done")
			done = true
		}
	}
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
