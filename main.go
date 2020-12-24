package main

import (
	"flag"

	imgui "github.com/AllenDang/giu/imgui"
)

const rootPathName string = "/Volumes/Work1/Musique/2020"

const (
	IDLE = iota
	RUNNING
	DONE

	GUI
	TUI
)

type AppCtx struct {
	scanRootPath string
	libRootPath  string
	scanResult   *ScanResult
	messages     []string
	font         imgui.Font
	scanStatus   int
}

func (ctx *AppCtx) scanStatusToString() string {
	switch ctx.scanStatus {
	case IDLE:
		return "IDLE"
	case RUNNING:
		return "Running"
	case DONE:
		return "Done"
	}
	return "Unknown status"
}

var appCtx AppCtx

func main() {
	noGUI := flag.Bool("no-gui", false, "disable GUI")
	flag.Parse()

	if *noGUI {
		startScan(TUI)
		//list()
	} else {
		setupGUI()
	}

}
