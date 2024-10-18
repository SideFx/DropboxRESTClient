// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// UI, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/assets"
	"Dropbox_REST_Client/models"
	"github.com/richardwilkes/unison"
)

const (
	wndMinWidth  float32 = 768
	wndMinHeight float32 = 480
)

var mainWindow *unison.Window
var mainContent *unison.Panel

func NewMainWindow() error {
	var err error
	mainWindow, err = unison.NewWindow(assets.AppName)
	if err != nil {
		return err
	}
	installDefaultMenus(mainWindow)
	loadSettings()
	mainContent = mainWindow.Content()
	mainContent.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(5)))
	mainContent.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: 1,
		VSpacing: 5,
	})
	mainContent.AddChild(createToolbarPanel())
	mainContent.AddChild(createTablePanel())
	mainWindow.Pack()
	// Set MainWindow size & position
	rect := _settings.WindowRect
	if rect.Width < wndMinWidth {
		rect.Width = wndMinWidth
	}
	if rect.Height < wndMinHeight {
		rect.Height = wndMinHeight
	}
	dispRect := unison.PrimaryDisplay().Usable
	if rect.X == 0 || rect.X > dispRect.Width-rect.Width {
		if dispRect.Width > rect.Width {
			rect.X = (dispRect.Width - rect.Width) / 2
		}
	}
	if rect.Y == 0 || rect.Y > dispRect.Height-rect.Height {
		if dispRect.Height > rect.Height {
			rect.Y = (dispRect.Height - rect.Height) / 2
		}
	}
	mainWindow.SetFrameRect(rect)
	installCallbacks()
	if IsTokenPresent() {
		models.DropboxReadRootFolders()
	}
	mainWindow.ToFront()
	return nil
}

func installCallbacks() {
	mainWindow.MinMaxContentSizeCallback = func() (minSize, maxSize unison.Size) {
		return windowMinMaxResizeCallback()
	}
	mainWindow.WillCloseCallback = func() {
		mainWindowWillClose()
	}
}

func windowMinMaxResizeCallback() (minSize, maxSize unison.Size) {
	var _min, _max unison.Size
	_min = unison.NewSize(wndMinWidth, wndMinHeight)
	disp := unison.PrimaryDisplay()
	_max.Width = disp.Usable.Width
	_max.Height = disp.Usable.Height
	return _min, _max
}

func mainWindowWillClose() {
	saveSettings()
}

func AllowQuitCallback() bool {
	mainWindow.AttemptClose()
	return true
}
