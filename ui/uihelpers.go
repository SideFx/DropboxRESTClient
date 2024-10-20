// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// UI utilities, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
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
var userInfoBtn *unison.Button
var refreshBtn *unison.Button
var addFolderBtn *unison.Button
var addRootFolderBtn *unison.Button
var deleteBtn *unison.Button
var uploadBtn *unison.Button
var downloadBtn *unison.Button
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
	userInfoBtn, err = createButton(assets.CapAboutUser, assets.IconUserInfo)
	if err == nil {
		userInfoBtn.SetEnabled(true)
		userInfoBtn.SetFocusable(false)
		panel.AddChild(userInfoBtn)
		userInfoBtn.ClickCallback = func() { aboutUser() }
	}
	createSpacer(30, panel)
	refreshBtn, err = createButton(assets.CapRefresh, assets.IconRefresh)
	if err == nil {
		refreshBtn.SetEnabled(true)
		refreshBtn.SetFocusable(false)
		panel.AddChild(refreshBtn)
		refreshBtn.ClickCallback = func() { refresh() }
	}
	addRootFolderBtn, err = createButton(assets.CapNewRootFolder, assets.IconAddFolder)
	if err == nil {
		addRootFolderBtn.SetEnabled(true)
		addRootFolderBtn.SetFocusable(false)
		panel.AddChild(addRootFolderBtn)
		addRootFolderBtn.ClickCallback = func() { newFolder(true) }
	}
	addFolderBtn, err = createButton(assets.CapNewFolder, assets.IconAddFolder)
	if err == nil {
		addFolderBtn.SetEnabled(true)
		addFolderBtn.SetFocusable(false)
		panel.AddChild(addFolderBtn)
		addFolderBtn.ClickCallback = func() { newFolder(false) }
	}
	deleteBtn, err = createButton(assets.CapDelete, assets.IconDelete)
	if err == nil {
		deleteBtn.SetEnabled(true)
		deleteBtn.SetFocusable(false)
		panel.AddChild(deleteBtn)
		deleteBtn.ClickCallback = func() { deleteItem() }
	}
	uploadBtn, err = createButton(assets.CapUpload, assets.IconUpload)
	if err == nil {
		uploadBtn.SetEnabled(true)
		uploadBtn.SetFocusable(false)
		panel.AddChild(uploadBtn)
		uploadBtn.ClickCallback = func() { uploadItems() }
	}
	downloadBtn, err = createButton(assets.CapDownload, assets.IconDownload)
	if err == nil {
		downloadBtn.SetEnabled(true)
		downloadBtn.SetFocusable(false)
		panel.AddChild(downloadBtn)
		downloadBtn.ClickCallback = func() { downloadItems() }
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
	content.AddChild(tableScrollArea)
}

func createSpacer(width float32, panel *unison.Panel) {
	spacer := &unison.Panel{}
	spacer.Self = spacer
	spacer.SetSizer(func(_ unison.Size) (minSize, prefSize, maxSize unison.Size) {
		minSize.Width = width
		prefSize.Width = width
		maxSize.Width = width
		return
	})
	panel.AddChild(spacer)
}
