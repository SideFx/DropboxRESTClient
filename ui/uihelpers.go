// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// UI utilities, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"Dropbox_REST_Client/dialogs"
	"Dropbox_REST_Client/models"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

const (
	toolbuttonWidth          = 20
	toolbuttonHeight         = 20
	toolbarFontSize  float32 = 9
)

var settingsBtn *unison.Button
var testBtn *unison.Button
var tableContent *unison.Panel

func newSVGButton(svg *unison.SVG) *unison.Button {
	btn := unison.NewButton()
	btn.HideBase = true
	btn.Drawable = &unison.DrawableSVG{
		SVG:  svg,
		Size: unison.NewSize(toolbuttonWidth, toolbuttonHeight),
	}
	btn.Font = unison.LabelFont.Face().Font(toolbarFontSize)
	return btn
}

func createButton(title string, svgcontent string) (*unison.Button, error) {
	svg, err := unison.NewSVGFromContentString(svgcontent)
	if err != nil {
		return nil, err
	}
	btn := newSVGButton(svg)
	btn.SetTitle(title)
	btn.SetLayoutData(align.Middle)
	return btn, nil
}

func createToolbarPanel() *unison.Panel {
	var err error
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlowLayout{
		HSpacing: 1,
		VSpacing: unison.StdVSpacing,
	})
	settingsBtn, err = createButton(assets.CapSettings, assets.IconSettings)
	if err == nil {
		settingsBtn.SetEnabled(true)
		settingsBtn.SetFocusable(false)
		panel.AddChild(settingsBtn)
		settingsBtn.ClickCallback = func() { SettingsDialog() }
	}
	testBtn, err = createButton(assets.CapAboutUser, assets.IconUserInfo)
	if err == nil {
		testBtn.SetEnabled(true)
		testBtn.SetFocusable(false)
		panel.AddChild(testBtn)
		testBtn.ClickCallback = func() { aboutUser() }
	}
	return panel
}

func installDefaultMenus(wnd *unison.Window) {
	unison.DefaultMenuFactory().BarForWindow(wnd, func(m unison.Menu) {
		unison.InsertStdMenus(m, dialogs.AboutDialog, SettingsDialogFromMenu, nil)
	})
}

func createTablePanel() *unison.Panel {
	tableContent = unison.NewPanel()
	tableContent.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: 1,
		VSpacing: 1,
	})
	tableContent.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	tableContent.SetBorder(unison.NewDefaultFieldBorder(false))
	newFileSystemTable(tableContent)
	return tableContent.AsPanel()
}

func newFileSystemTable(content *unison.Panel) {
	table, header := models.NewFileSystemTable()
	header.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	tableScrollArea := unison.NewScrollPanel()
	tableScrollArea.SetContent(table, behavior.Fill, behavior.Fill)
	tableScrollArea.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	tableScrollArea.SetColumnHeader(header)
	table.SelectionChangedCallback = func() {

	}
	content.AddChild(tableScrollArea)
}

func aboutUser() {
	userinfo, err := api.GetCurrentUser()
	if err == nil {
		dialogs.AboutUserDialog(userinfo)
	}
}
