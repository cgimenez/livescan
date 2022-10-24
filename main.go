package main

import (
	"database/sql"
	"flag"
	"fmt"

	imgui "github.com/AllenDang/giu/imgui"
)

const (
	IDLE = iota
	RUNNING
	DONE

	STATE_PACKS
	STATE_SCAN

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
	dbCnx         *sql.DB
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

func setAppCtxDefaults() {
	var err error

	appCtx.scanRootPath = "/Volumes/Work1/Musique/Génériques Pour la Nouvelle Télévision/Compos/Bloc 4"
	appCtx.libRootPath = "/Volumes/Work1/Musique/Sound banks/Packs Live"
	appCtx.deleteEnabled = false
	appCtx.fileExistsFn = fileExists
	if appCtx.dbCnx, err = newDBCnx(); err != nil {
		panic(err)
	}
}

var appCtx AppCtx

func main() {
	setAppCtxDefaults()
	defer appCtx.dbCnx.Close()

	runMode := flag.String("runmode", "", "tui, debug or nothing")
	flag.Parse()

	switch *runMode {
	case "tui":
		startScan(TUI)
		appCtx.scanResult.list()
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
