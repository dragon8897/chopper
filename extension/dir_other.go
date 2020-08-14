// +build !windows,!android,!ios

package extension

import (
	"fyne.io/fyne"
	"fyne.io/fyne/theme"
)

func (f *dirDialog) loadPlaces() []fyne.CanvasObject {
	return []fyne.CanvasObject{makeFavoriteButton("Computer", theme.ComputerIcon(), func() {
		f.setDirectory("/")
	})}
}

func isHidden(file, _ string) bool {
	return len(file) == 0 || file[0] == '.'
}
