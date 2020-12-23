package main

import (
	"time"

	"github.com/AllenDang/giu"
	g "github.com/AllenDang/giu"
	imgui "github.com/AllenDang/giu/imgui"
)

var (
	messages   []string
	font       imgui.Font
	isScanning bool
)

func loadFont() {
	fonts := giu.Context.IO().Fonts()
	font = fonts.AddFontFromFileTTF("assets/arial.ttf", 13)
}

func onOpenFolder() {
}

func onStartScan() {
	if isScanning {
		return
	}
	go func() {
		isScanning = true
		messages = make([]string, 0)
		sc := newScanResult(rootPathName)

		messages = append(messages, "Finding files...à la tête de nël")
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

		isScanning = true
		for isScanning {
			msg, more := <-ch
			if more {
				messages = append(messages, msg)
				println(msg)
			} else {
				messages = append(messages, "done")
				isScanning = false
			}
			giu.Update()
		}
	}()
}

func loop() {
	g.PushFont(font)
	g.SingleWindow("ALPS", g.Layout{
		g.Line(
			g.Button("Open folder", onOpenFolder),
			g.Button("Start scan", onStartScan),
		),
		g.Child("Top panel", true, -1, 200, g.WindowFlagsHorizontalScrollbar, g.Layout{
			g.Custom(func() {
				for _, msg := range messages {
					//g.Label(msg).Build()
					imgui.Text(msg)
				}
				//g.SetScrollHereY(1.0)
			}),
		}),
		g.Custom(func() {
			if time.Now().Second()%2 == 0 {
				// Enter 'Top panel' child and call SetScrollHereY()
				imgui.BeginChild("Top panel")
				imgui.SetScrollHereY(1)
				imgui.EndChild()
			}
		}),
		g.Child("Bottom panel", true, -1, 200, 0, g.Layout{}),
	})
	g.PopFont()
}

func setupGUI() {
	wnd := g.NewMasterWindow("Ableton Live Projects Scanner", 800, 500, g.MasterWindowFlagsNotResizable, loadFont)
	wnd.Main(loop)
}
