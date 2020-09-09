package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/dragon8897/chopper/extension"
)

const preferenceChopperCfg = "chopperCfg"

type ChopperCfg struct {
	ID      int64  `json:"id"`
	DirPath string `json:"dir"`
	Git     struct {
		URL      string `json:"url"`
		UserName string `json:"name"`
		Password string `json:"pwd"`
	} `json:"git"`
	Robot struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	} `json:"robot"`
}

var allCfg []ChopperCfg
var chopperPanel *widget.Box

func newChopperCfg(win fyne.Window) {
	cfg := ChopperCfg{
		ID: time.Now().UnixNano(),
	}
	allCfg = append(allCfg, cfg)
	chopperPanel.Append(createCfgUI(&cfg, win))
}

func deleteChopperCfg(id int64) {
	var index = -1
	count := len(allCfg)
	for i := 0; i < count; i++ {
		cfg := allCfg[i]
		if cfg.ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return
	}
	if index >= count {
		return
	}
	allCfg = append(allCfg[:index], allCfg[index+1:]...)
	children := chopperPanel.Children
	chopperPanel.Children = append(children[:index], children[index+1:]...)
	chopperPanel.Refresh()
}

func createCfgUI(cfg *ChopperCfg, win fyne.Window) fyne.CanvasObject {
	content := widget.NewEntry()
	content.PlaceHolder = "请输入任务完成后机器人的发送内容"
	content.Text = cfg.Robot.Content
	content.OnChanged = func(text string) {
		cfg.Robot.Content = text
	}

	robots := widget.NewSelect([]string{"游戏开发"}, func(name string) {
		cfg.Robot.Name = name
	})
	robots.PlaceHolder = "请先选择一个群聊机器人"
	robots.Selected = cfg.Robot.Name

	btnDir := &widget.Button{}
	btnDir.Alignment = widget.ButtonAlignLeading
	if cfg.DirPath == "" {
		btnDir.Text = "点击这里输入图片导出的文件夹"
	} else {
		btnDir.Text = cfg.DirPath
	}
	btnDir.OnTapped = func() {
		extension.ShowDirSelect(func(dirPath string, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			cfg.DirPath = dirPath
			btnDir.Text = dirPath
			btnDir.Refresh()
		}, win)

	}
	btnDirRow := fyne.NewContainerWithLayout(layout.NewFormLayout(), []fyne.CanvasObject{
		widget.NewLabel("图片目录:"),
		btnDir,
	}...)

	entryURL := widget.NewEntry()
	entryURL.PlaceHolder = "请输入 git 地址"
	entryURL.Text = cfg.Git.URL
	entryURL.OnChanged = func(text string) {
		cfg.Git.URL = text
	}
	entryURLRow := fyne.NewContainerWithLayout(layout.NewFormLayout(), []fyne.CanvasObject{
		widget.NewLabel("地址:"),
		entryURL,
	}...)

	entryGitName := widget.NewEntry()
	entryGitName.PlaceHolder = "请输入 git 账号"
	entryGitName.Text = cfg.Git.UserName
	entryGitName.OnChanged = func(text string) {
		cfg.Git.UserName = text
	}
	entryGitNameRow := fyne.NewContainerWithLayout(layout.NewFormLayout(), []fyne.CanvasObject{
		widget.NewLabel("账号:"),
		entryGitName,
	}...)

	entryGitPwd := widget.NewPasswordEntry()
	entryGitPwd.PlaceHolder = "请输入 git 密码"
	entryGitPwd.Text = cfg.Git.Password
	entryGitPwd.OnChanged = func(text string) {
		cfg.Git.Password = text
	}
	entryGitPwdRow := fyne.NewContainerWithLayout(layout.NewFormLayout(), []fyne.CanvasObject{
		widget.NewLabel("密码:"),
		entryGitPwd,
	}...)

	btnStart := widget.NewButton("      开始      ", func() {
		export(*cfg, win)
	})

	btnStart.Style = widget.PrimaryButton

	return widget.NewVBox(
		layout.NewSpacer(),
		widget.NewGroup(" ", layout.NewSpacer()),
		widget.NewHBox(
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
				deleteChopperCfg(cfg.ID)
			}),
		),
		btnDirRow,
		widget.NewAccordionContainer(
			widget.NewAccordionItem("Git 配置",
				widget.NewVBox(
					entryURLRow,
					entryGitNameRow,
					entryGitPwdRow,
				),
			),
			widget.NewAccordionItem("机器人配置",
				widget.NewVBox(
					robots,
					content,
				),
			),
		),
		widget.NewHBox(
			layout.NewSpacer(),
			btnStart,
		),
		layout.NewSpacer(),
	)
}

func chopperScreen(a fyne.App, win fyne.Window) fyne.CanvasObject {
	btnAdd := widget.NewButtonWithIcon("新建", theme.ContentAddIcon(), func() {
		newChopperCfg(win)
	})

	var allCanvas []fyne.CanvasObject
	for index := range allCfg {
		allCanvas = append(allCanvas, createCfgUI(&allCfg[index], win))
	}
	chopperPanel = widget.NewVBox(allCanvas...)

	content := []fyne.CanvasObject{
		widget.NewLabelWithStyle("欢迎使用图片资源命名工具", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		chopperPanel,
		layout.NewSpacer(),
		widget.NewHBox(
			layout.NewSpacer(),
			btnAdd,
			layout.NewSpacer(),
		),
		layout.NewSpacer(),
	}

	return widget.NewVBox(content...)
}

func main() {
	os.Setenv("FYNE_FONT", "/Library/Fonts/Arial Unicode.ttf")
	defer os.Unsetenv("FYNE_FONT")
	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(theme.FyneLogo())

	cfg := a.Preferences().String(preferenceChopperCfg)
	fmt.Println("chopper cfg", cfg)
	if len(cfg) > 0 {
		err := json.Unmarshal([]byte(cfg), &allCfg)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		cfgStr, err := json.Marshal(&allCfg)
		if err == nil {
			a.Preferences().SetString(preferenceChopperCfg, string(cfgStr))
		}
	}()

	w := a.NewWindow("Chopper")
	w.SetMaster()
	w.Resize(fyne.NewSize(500, 300))
	w.SetContent(chopperScreen(a, w))

	w.ShowAndRun()
}
