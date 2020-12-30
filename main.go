package main

import (
	"flag"
	"fmt"

	imgui "github.com/AllenDang/giu/imgui"
)

const rootPathName string = "/Volumes/Work1/Musique/2019-Curated-2"

const (
	IDLE = iota
	RUNNING
	DONE

	GUI
	TUI

	DEVELOP
	TEST
)

type AppCtx struct {
	scanRootPath  string
	libRootPath   string
	scanResult    *ScanResult
	messages      []string
	font, icons   imgui.Font
	scanStatus    int
	deleteEnabled bool
	env           int
	fileExistsFn  func(string) bool
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
	appCtx.scanRootPath = rootPathName
	appCtx.libRootPath = "/Volumes/Work1/Musique/Sound banks/Packs Live"
	appCtx.deleteEnabled = false
	appCtx.fileExistsFn = fileExists
	runMode := flag.String("runmode", "", "tui, debug or nothing")
	flag.Parse()

	switch *runMode {
	case "tui":
		startScan(TUI)
		list()
	case "debug":
		appCtx.scanResult = newScanResult(appCtx.scanRootPath)
		appCtx.scanResult.debug()
	case "scanpacks":
		_, err := isLivePacksDBScanned()
		if err != nil {
			panic(err)
		}
		fmt.Println("Will scan Live Packs directory")
		if err = scanLivePacks(); err != nil {
			panic(err)
		}
	default:
		setupGUI()
	}

}
