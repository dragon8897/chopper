package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/cmd/fyne_demo/data"
	"fyne.io/fyne/cmd/fyne_demo/screens"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/dragon8897/chopper/extension"
)

const preferenceCurrentTab = "currentTab"
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

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

// func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
// 	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
// 		focused.TypedShortcut(s)
// 	}
// }

func welcomeScreen(a fyne.App, win fyne.Window) fyne.CanvasObject {
	logo := canvas.NewImageFromResource(data.FyneScene)
	if fyne.CurrentDevice().IsMobile() {
		logo.SetMinSize(fyne.NewSize(171, 125))
	} else {
		logo.SetMinSize(fyne.NewSize(228, 167))
	}

	content := widget.NewEntry()
	content.PlaceHolder = "请输入任务完成后机器人的发送内容"

	robots := widget.NewSelect([]string{"机器人 1", "机器人 2", "机器人 3"}, func(s string) { fmt.Println("selected", s) })
	robots.PlaceHolder = "请先选择一个群聊机器人"

	return widget.NewVBox(
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Welcome to the Fyne toolkit demo app", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewHBox(layout.NewSpacer(), logo, layout.NewSpacer()),

		widget.NewHBox(layout.NewSpacer(),
			widget.NewHyperlink("fyne.io", parseURL("https://fyne.io/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("documentation", parseURL("https://fyne.io/develop/")),
			widget.NewLabel("-"),
			widget.NewHyperlink("sponsor", parseURL("https://github.com/sponsors/fyne-io")),
			layout.NewSpacer(),
		),
		layout.NewSpacer(),

		widget.NewGroup("Theme",
			fyne.NewContainerWithLayout(layout.NewGridLayout(2),
				widget.NewButton("Dark", func() {
					a.Settings().SetTheme(theme.DarkTheme())
				}),
				widget.NewButton("Light", func() {
					a.Settings().SetTheme(theme.LightTheme())
				}),
			),
		),
	)
}

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

	robots := widget.NewSelect([]string{"机器人 1", "机器人 2", "机器人 3"}, func(name string) {
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

	return widget.NewVBox(
		layout.NewSpacer(),
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
			widget.NewButton("开始", func() {
				export(*cfg, win)
			}),
		),
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

	// newItem := fyne.NewMenuItem("New", nil)
	// otherItem := fyne.NewMenuItem("Other", nil)
	// otherItem.ChildMenu = fyne.NewMenu("",
	// 	fyne.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
	// 	fyne.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") }),
	// )
	// newItem.ChildMenu = fyne.NewMenu("",
	// 	fyne.NewMenuItem("File", func() { fmt.Println("Menu New->File") }),
	// 	fyne.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") }),
	// 	otherItem,
	// )
	// settingsItem := fyne.NewMenuItem("Settings", func() { fmt.Println("Menu Settings") })

	// cutItem := fyne.NewMenuItem("Cut", func() {
	// 	shortcutFocused(&fyne.ShortcutCut{
	// 		Clipboard: w.Clipboard(),
	// 	}, w)
	// })
	// copyItem := fyne.NewMenuItem("Copy", func() {
	// 	shortcutFocused(&fyne.ShortcutCopy{
	// 		Clipboard: w.Clipboard(),
	// 	}, w)
	// })
	// pasteItem := fyne.NewMenuItem("Paste", func() {
	// 	shortcutFocused(&fyne.ShortcutPaste{
	// 		Clipboard: w.Clipboard(),
	// 	}, w)
	// })
	// findItem := fyne.NewMenuItem("Find", func() { fmt.Println("Menu Find") })

	// helpMenu := fyne.NewMenu("Help", fyne.NewMenuItem("Help", func() { fmt.Println("Help Menu") }))
	// mainMenu := fyne.NewMainMenu(
	// 	// a quit item will be appended to our first menu
	// 	fyne.NewMenu("File", newItem, fyne.NewMenuItemSeparator(), settingsItem),
	// 	fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
	// 	helpMenu,
	// )
	// w.SetMainMenu(mainMenu)
	w.SetMaster()

	tabs := widget.NewTabContainer(
		widget.NewTabItemWithIcon("Chopper", theme.ConfirmIcon(), chopperScreen(a, w)),
		widget.NewTabItemWithIcon("Welcome", theme.HomeIcon(), welcomeScreen(a, w)),
		widget.NewTabItemWithIcon("Graphics", theme.DocumentCreateIcon(), screens.GraphicsScreen()),
		widget.NewTabItemWithIcon("Widgets", theme.CheckButtonCheckedIcon(), screens.WidgetScreen()),
		widget.NewTabItemWithIcon("Containers", theme.ViewRestoreIcon(), screens.ContainerScreen()),
		widget.NewTabItemWithIcon("Windows", theme.ViewFullScreenIcon(), screens.DialogScreen(w)))

	if !fyne.CurrentDevice().IsMobile() {
		tabs.Append(widget.NewTabItemWithIcon("Advanced", theme.SettingsIcon(), screens.AdvancedScreen(w)))
	}
	tabs.SetTabLocation(widget.TabLocationLeading)
	tabs.SelectTabIndex(a.Preferences().Int(preferenceCurrentTab))
	w.SetContent(tabs)

	w.ShowAndRun()
	a.Preferences().SetInt(preferenceCurrentTab, tabs.CurrentTabIndex())
}
