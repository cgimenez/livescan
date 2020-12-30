package main

import (
	"fmt"
	"time"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/giu/imgui"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sqweek/dialog"
)

func loadFont() {
	fonts := giu.Context.IO().Fonts()
	appCtx.font = fonts.AddFontFromFileTTF("assets/Karla-Regular.ttf", 13)
	appCtx.icons = fonts.AddFontFromFileTTF("assets/Font-Awesome-5-Free-Solid-900.ttf", 13)
}

func onOpenFolder() {
	d, _ := dialog.Directory().Title("Now find a dir").Browse()
	appCtx.scanRootPath = d
	appCtx.scanStatus = IDLE
}

func onStartScan() {
	if appCtx.scanStatus == RUNNING {
		return
	}
	go func() {
		startScan(GUI)
	}()
}

func Button(label string, width float32, height float32, clickable bool, callback func()) {
	if !clickable {
		imgui.PushStyleColor(imgui.StyleColorButton, RGBA(153, 153, 153, 255))
	}
	if imgui.ButtonV(label, imgui.Vec2{X: width, Y: height}) && clickable {
		callback()
	}
	if !clickable {
		imgui.PopStyleColor()
	}
}

func scanOutput() {
	w, _ := glfw.GetCurrentContext().GetSize()
	imgui.BeginChildV("messages", imgui.Vec2{X: float32(w - 20), Y: 200}, true, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsAlwaysVerticalScrollbar)
	if appCtx.scanStatus == RUNNING || appCtx.scanStatus == DONE {
		for _, msg := range appCtx.messages {
			imgui.Text(msg)
		}
	}
	if appCtx.scanStatus == RUNNING && time.Now().Second()%2 == 0 {
		imgui.SetScrollHereY(1)
	}
	imgui.EndChild()
}

func liveFilesTab() {
	if imgui.BeginTabItem("Live Projects") {
		if appCtx.scanStatus == DONE {
			for _, liveFilesKey := range appCtx.scanResult.liveFilesKeys {
				liveFile := appCtx.scanResult.liveFiles[liveFilesKey]
				if imgui.TreeNode(liveFile.pathname) {
					imgui.SetNextItemOpen(true, imgui.ConditionFirstUseEver)
					if imgui.TreeNode("Refs") {
						for _, ref := range liveFile.refs {
							imgui.Text(ref.pathname)
						}
						imgui.TreePop()
					}
					imgui.SetNextItemOpen(true, imgui.ConditionFirstUseEver)
					if imgui.TreeNode("Outsides") {
						for _, libref := range liveFile.outsiderefs {
							imgui.Text("Outside : " + libref)
						}
						imgui.TreePop()
					}
					imgui.SetNextItemOpen(true, imgui.ConditionFirstUseEver)
					if imgui.TreeNode("Missing") {
						for _, missing := range liveFile.missing {
							imgui.Text("Missing : " + missing)
						}
						imgui.TreePop()
					}
					imgui.TreePop()
				}
			}
		}
		imgui.EndTabItem()
	}
}

func buildRows() []*g.RowWidget {
	rows := make([]*g.RowWidget, 0)

	appCtx.scanResult.audioFilesSize = 0
	for i, audiofileKey := range appCtx.scanResult.audioFilesKeys {
		audioFile := appCtx.scanResult.audioFiles[audiofileKey]
		if len(audioFile.refs) == 0 {
			appCtx.scanResult.audioFilesSize += audioFile.size
			row := g.Row(
				g.Label(audioFile.pathname),
				g.Custom(func() {
					imgui.PushID(fmt.Sprintf("%d", i))
					if appCtx.deleteEnabled {
						g.ButtonV("Delete", 100, 20, func() {}).Build()
					}
					imgui.PopID()
				}),
			)
			rows = append(rows, row)
		}
	}
	return rows
}

func audioFilesTab() {
	if imgui.BeginTabItem(fmt.Sprintf("Orphan audio files  [ %s ]", ByteCountSI(appCtx.scanResult.audioFilesSize))) {
		if appCtx.scanStatus == DONE {
			if len(appCtx.scanResult.audioFilesKeys) == 0 {
				imgui.Text("No orphans found.")
			} else {
				g.Child("Container", true, 0, 0, 0, g.Layout{
					g.FastTable("Fast table", true, buildRows()),
				}).Build()
			}
		}
		imgui.EndTabItem()
	}
}

func loop() {
	g.PushFont(appCtx.font)
	var showWindow bool = true

	imgui.PushStyleVarFloat(imgui.StyleVarFrameRounding, 0)
	style := imgui.CurrentStyle()
	style.SetColor(imgui.StyleColorText, RGBA(0, 0, 0, 255))
	style.SetColor(imgui.StyleColorButton, RGBA(183, 183, 183, 255))
	style.SetColor(imgui.StyleColorBorder, RGBA(200, 200, 200, 255))
	style.SetColor(imgui.StyleColorWindowBg, RGBA(118, 118, 118, 255))
	style.SetColor(imgui.StyleColorChildBg, RGBA(181, 178, 177, 255))

	style.SetColor(imgui.StyleColorTab, RGBA(225, 171, 0, 255))
	style.SetColor(imgui.StyleColorTabActive, RGBA(255, 201, 7, 255))

	style.SetColor(imgui.StyleColorScrollbarBg, RGBA(140, 140, 140, 255))
	style.SetColor(imgui.StyleColorScrollbarGrab, RGBA(52, 55, 57, 255))
	style.SetColor(imgui.StyleColorScrollbarGrabHovered, RGBA(52, 55, 57, 255))

	w, h := glfw.GetCurrentContext().GetSize()
	imgui.SetNextWindowPosV(imgui.Vec2{X: 0, Y: 0}, imgui.ConditionFirstUseEver, imgui.Vec2{X: 0, Y: 0})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: float32(w), Y: float32(h)}, imgui.ConditionFirstUseEver)
	imgui.BeginV("ALPS", &showWindow, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoDecoration)

	delBtnCaption := ""
	if appCtx.deleteEnabled {
		delBtnCaption = "Disable delete"
	} else {
		delBtnCaption = "Enable delete"
	}
	Button(delBtnCaption, 100, 20, true, func() {
		appCtx.deleteEnabled = !appCtx.deleteEnabled
	})

	imgui.SameLine()
	Button("Open Folder", 100, 20, appCtx.scanStatus != RUNNING, onOpenFolder)
	imgui.SameLine()
	Button("Start scan", 100, 20, appCtx.scanStatus != RUNNING, onStartScan)
	imgui.SameLine()
	Button("Export to txt", 100, 20, appCtx.scanStatus == DONE, nil)

	imgui.Text("Status : " + appCtx.scanStatusToString())
	imgui.SameLine()
	imgui.Bullet()
	imgui.SameLine()
	imgui.Text(appCtx.scanRootPath)

	scanOutput()

	imgui.BeginChildV("lifefiles", imgui.Vec2{X: float32(w - 20), Y: -1}, true, imgui.WindowFlagsNoCollapse|imgui.WindowFlagsNoDecoration|imgui.WindowFlagsNoMove|imgui.WindowFlagsAlwaysVerticalScrollbar)
	if imgui.BeginTabBarV("Tabs", imgui.TabBarFlagsNoTooltip) {
		liveFilesTab()
		audioFilesTab()
		imgui.EndTabBar()
	}
	imgui.EndChild()

	imgui.PopStyleVar() // StyleVarFrameRounding
	imgui.End()
	g.PopFont()
}

func setupGUI() {
	wnd := g.NewMasterWindow("Ableton Live Projects Scanner", 1200, 500, g.MasterWindowFlagsMaximized, loadFont)
	onStartScan()
	wnd.Main(loop)
}
