package main

import (
	"time"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/giu/imgui"
	"github.com/sqweek/dialog"
)

func loadFont() {
	fonts := giu.Context.IO().Fonts()
	appCtx.font = fonts.AddFontFromFileTTF("assets/arial.ttf", 13)
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

func loop() {
	g.PushFont(appCtx.font)
	g.SingleWindow("ALPS", g.Layout{
		g.Line(
			g.Button("Open folder", onOpenFolder),
			g.Button("Start scan", onStartScan),
			g.Label(" : "+appCtx.scanRootPath),
		),
		g.Label("Status : " + appCtx.scanStatusToString()),
		g.Child("Top panel", true, -1, 190, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.Custom(func() {
				if appCtx.scanStatus == RUNNING || appCtx.scanStatus == DONE {
					for _, msg := range appCtx.messages {
						imgui.Text(msg)
					}
				}
			}),
		}),
		g.Custom(func() {
			if time.Now().Second()%2 == 0 {
				imgui.BeginChild("Top panel")
				imgui.SetScrollHereY(1)
				imgui.EndChild()
			}
		}),
		g.Child("Bottom panel", true, -1, -1, 0, g.Layout{
			g.TabBar("Tab audiofiles", g.Layout{
				g.TabItem("Live Projects", g.Layout{
					g.Custom(displayLiveProjects),
				}),
				g.TabItem("Orphan audio files", g.Layout{
					g.Custom(displayAudioFiles),
				}),
			}),
		}),
	})
	g.PopFont()
}

func displayAudioFiles() {
	if appCtx.scanStatus == DONE {
		if len(appCtx.scanResult.audioFilesKeys) == 0 {
			imgui.Text("No orphans found.")
			return
		}

		for _, audiofileKey := range appCtx.scanResult.audioFilesKeys {
			audioFile := appCtx.scanResult.audioFiles[audiofileKey]
			if len(audioFile.refs) == 0 {
				imgui.Text(audioFile.pathname)
				//g.Label(audiofile.pathname).Build()
			}
		}
	}
}

func displayLiveProjects() {
	if appCtx.scanStatus == DONE {
		if len(appCtx.scanResult.liveFilesKeys) == 0 {
			imgui.Text("No Live projects found.")
			return
		}

		for _, liveFilesKey := range appCtx.scanResult.liveFilesKeys {
			liveFile := appCtx.scanResult.liveFiles[liveFilesKey]
			g.TreeNode(liveFile.pathname, g.TreeNodeFlagsCollapsingHeader, g.Layout{
				g.Custom(func() {
					//println(len(liveFile.refs))
					//g.Selectable("Aie !", func() { fmt.Println(1) })
					//g.Button("Button inside tree", nil)
					//g.Label("Audio files refs")
					for _, ref := range liveFile.refs {
						imgui.Text("Ref :")
						imgui.SameLine()
						imgui.Text(ref.pathname)
						//g.Label(ref.pathname)
					}
					for _, missing := range liveFile.missing {
						imgui.Text("Missing: " + missing)
					}
					for _, libref := range liveFile.librefs {
						imgui.Text("Library: " + libref)
					}
				}),
				//g.Label(liveFile.pathname),
				//g.Selectable(liveFile.pathname, func() { fmt.Println(1) })
				//g.Button("Button inside tree", nil),
			}).Build()
		}
	}
}

func setupGUI() {
	wnd := g.NewMasterWindow("Ableton Live Projects Scanner", 800, 500, g.MasterWindowFlagsMaximized, loadFont)
	onStartScan()
	wnd.Main(loop)
}
