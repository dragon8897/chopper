package main

import (
	"fmt"
	"net/url"
	"os"

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

func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

func shortcutFocused(s fyne.Shortcut, w fyne.Window) {
	if focused, ok := w.Canvas().Focused().(fyne.Shortcutable); ok {
		focused.TypedShortcut(s)
	}
}

func makeFormTab() fyne.Widget {
	name := widget.NewEntry()
	name.SetPlaceHolder("John Smith")
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "账号:", Widget: name},
			{Text: "密码:", Widget: password},
		},
		OnCancel: func() {
			fmt.Println("Cancelled")
		},
		OnSubmit: func() {
			fmt.Println("Form submitted")
		},
	}
	return form
}

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
			widget.NewButton("open dialog", func() {
				fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
					if err == nil && reader == nil {
						return
					}
					if err != nil {
						dialog.ShowError(err, win)
						return
					}

					// fileOpened(reader)
				}, win)
				// fd.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".txt"}))
				fd.Show()
			}),
			widget.NewButton("open dir", func() {
				extension.ShowDirSelect(func(dirPath string, err error) {
					if err != nil {
						dialog.ShowError(err, win)
						return
					}

					fmt.Println("open dir:", dirPath)
				}, win)
			}),
			layout.NewSpacer(),
		),
		layout.NewSpacer(),

		widget.NewAccordionContainer(
			widget.NewAccordionItem("Git 配置",
				makeFormTab(),
			),
			widget.NewAccordionItem("机器人配置",
				widget.NewVBox(
					robots,
					content,
				),
			),
		),

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

func main() {
	os.Setenv("FYNE_FONT", "/Library/Fonts/Arial Unicode.ttf")
	defer os.Unsetenv("FYNE_FONT")
	a := app.NewWithID("io.fyne.demo")
	a.SetIcon(theme.FyneLogo())

	w := a.NewWindow("Fyne Demo")

	newItem := fyne.NewMenuItem("New", nil)
	otherItem := fyne.NewMenuItem("Other", nil)
	otherItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("Project", func() { fmt.Println("Menu New->Other->Project") }),
		fyne.NewMenuItem("Mail", func() { fmt.Println("Menu New->Other->Mail") }),
	)
	newItem.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem("File", func() { fmt.Println("Menu New->File") }),
		fyne.NewMenuItem("Directory", func() { fmt.Println("Menu New->Directory") }),
		otherItem,
	)
	settingsItem := fyne.NewMenuItem("Settings", func() { fmt.Println("Menu Settings") })

	cutItem := fyne.NewMenuItem("Cut", func() {
		shortcutFocused(&fyne.ShortcutCut{
			Clipboard: w.Clipboard(),
		}, w)
	})
	copyItem := fyne.NewMenuItem("Copy", func() {
		shortcutFocused(&fyne.ShortcutCopy{
			Clipboard: w.Clipboard(),
		}, w)
	})
	pasteItem := fyne.NewMenuItem("Paste", func() {
		shortcutFocused(&fyne.ShortcutPaste{
			Clipboard: w.Clipboard(),
		}, w)
	})
	findItem := fyne.NewMenuItem("Find", func() { fmt.Println("Menu Find") })

	helpMenu := fyne.NewMenu("Help", fyne.NewMenuItem("Help", func() { fmt.Println("Help Menu") }))
	mainMenu := fyne.NewMainMenu(
		// a quit item will be appended to our first menu
		fyne.NewMenu("File", newItem, fyne.NewMenuItemSeparator(), settingsItem),
		fyne.NewMenu("Edit", cutItem, copyItem, pasteItem, fyne.NewMenuItemSeparator(), findItem),
		helpMenu,
	)
	w.SetMainMenu(mainMenu)
	w.SetMaster()

	tabs := widget.NewTabContainer(
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
