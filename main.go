package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gotk3/gotk3/gtk"
)

func displayMsg(msg string) {
	log.Println(msg)
}

func startScan() {
	files, err := scanDirectory("/Volumes/Work1/Musique/Compositions Live")
	if err != nil {
		panic(err)
	}

	if err = scanLiveProjects(&files, displayMsg); err != nil {
		panic(err)
	}

	log.Println("Orphan audio files")
	var sz int64
	for _, audioFile := range files.audioFiles {
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

func tvAddText(tv *gtk.TextView, text string) {
	buffer, _ := tv.GetBuffer()
	start, end := buffer.GetBounds()
	t, _ := buffer.GetText(start, end, true)
	buffer.SetText(t + text)
}

func setupGUI() {
	gtk.Init(nil)
	b, err := gtk.BuilderNew()
	if err != nil {
		log.Fatal(err)
	}
	err = b.AddFromFile("ui.glade")
	if err != nil {
		log.Fatal(err)
	}

	obj, err := b.GetObject("window_main")
	if err != nil {
		log.Fatal(err)
	}
	win := obj.(*gtk.Window)
	_, _ = win.Connect("destroy", func() {
		gtk.MainQuit()
	})

	obj, _ = b.GetObject("btn_start_scan")
	btn_start_scan := obj.(*gtk.Button)
	btn_start_scan.SetSensitive(false)

	obj, _ = b.GetObject("text_view")
	text_output := obj.(*gtk.TextView)
	text_output.SetMonospace(true)

	_, _ = btn_start_scan.Connect("clicked", func() {
		files, _ := scanDirectory("/Volumes/Work1/Musique/Compositions Live/01-2019")
		err = scanLiveProjects(&files, func(msg string) {
			for gtk.EventsPending() {
				gtk.MainIteration()
			}
			tvAddText(text_output, msg+"\n")
		})
		if err != nil {
			tvAddText(text_output, err.Error())
		}
	})

	win.ShowAll()
	gtk.Main()
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
