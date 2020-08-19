package main

import (
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
)

func export(cfg ChopperCfg, win fyne.Window) {
	f, err := os.Stat(cfg.DirPath)
	if err != nil {
		return
	}
	if !f.IsDir() {
		return
	}
	prog := dialog.NewProgressInfinite("导出", "正在导出", win)
	prog.Show()

	go func() {
		time.Sleep(time.Second * 5)
		prog.Hide()
	}()

}
