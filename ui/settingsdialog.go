// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// App auhtorization dialog, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"fmt"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const inpTextSizeMax = 200
const obscureRune = 0x2a

var okButton *unison.Button
var authButton *unison.Button
var inpAppKey *unison.Field
var inpAppSecret *unison.Field
var inpAuthCode *unison.Field
var authorizeSucceeded = false

func SettingsDialogFromMenu(_ unison.MenuItem) {
	SettingsDialog()
}

func newUserButtonInfo() *unison.DialogButtonInfo {
	return &unison.DialogButtonInfo{
		Title:        assets.CapAboutUser,
		ResponseCode: unison.ModalResponseUserBase,
		KeyCodes:     []unison.KeyCode{unison.KeyLControl + unison.KeyA},
	}
}

func SettingsDialog() {
	dialog, err := unison.NewDialog(nil, nil, newPreferencesPanel(),
		[]*unison.DialogButtonInfo{unison.NewOKButtonInfo(), newUserButtonInfo(), unison.NewCancelButtonInfo()},
		unison.NotResizableWindowOption())
	if err == nil {
		wnd := dialog.Window()
		wnd.SetTitle(assets.CapSettings)
		//prepareTitleIcon()
		//if len(titleIcons) > 0 {
		//	wnd.SetTitleIcons(titleIcons)
		//}
		okButton = dialog.Button(unison.ModalResponseOK)
		okButton.ClickCallback = func() {
			saveSettings()
			dialog.StopModal(unison.ModalResponseOK)
		}
		authButton = dialog.Button(unison.ModalResponseUserBase)
		authButton.ClickCallback = func() {
			getAuthorizationCode()
		}
		_ = dialog.Button(unison.ModalResponseCancel)
		inpAppKey.SetText(_settings.AppAuth.AppKey)
		inpAppSecret.SetText(_settings.AppAuth.AppSecret)
		okButton.SetEnabled(checkOk())
		okButton = dialog.Button(unison.ModalResponseOK)
		okButton.ClickCallback = func() {
			save()
			dialog.StopModal(unison.ModalResponseOK)
		}
		inpAuthCode.SetEnabled(false)
		dialog.RunModal()
	}
}

func newPreferencesPanel() *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	lblAppKey := unison.NewLabel()
	lblAppKey.Font = unison.LabelFont
	lblAppKey.SetTitle(assets.CapAppKey)
	inpAppKey = unison.NewField()
	inpAppKey.Font = unison.FieldFont
	inpAppKey.MinimumTextWidth = inpTextSizeMax
	inpAppKey.ObscurementRune = obscureRune
	lblAppSecret := unison.NewLabel()
	lblAppSecret.Font = unison.LabelFont
	lblAppSecret.SetTitle(assets.CapAppSecret)
	inpAppSecret = unison.NewField()
	inpAppSecret.Font = unison.FieldFont
	inpAppSecret.MinimumTextWidth = inpTextSizeMax
	inpAppSecret.ObscurementRune = obscureRune
	lblAuthCode := unison.NewLabel()
	lblAuthCode.Font = unison.LabelFont
	lblAuthCode.SetTitle(assets.CapAuthorizationCode)
	inpAuthCode = unison.NewField()
	inpAuthCode.Font = unison.FieldFont
	inpAuthCode.MinimumTextWidth = inpTextSizeMax
	inpAuthCode.ObscurementRune = obscureRune
	inpAppKey.ModifiedCallback = func(before, after *unison.FieldState) {
		inpModifiedCallback(before, after)
	}
	inpAppSecret.ModifiedCallback = func(before, after *unison.FieldState) {
		inpModifiedCallback(before, after)
	}
	inpAuthCode.ModifiedCallback = func(before, after *unison.FieldState) {
		inpModifiedCallback(before, after)
	}
	panel.SetLayoutData(&unison.FlexLayoutData{
		MinSize: unison.Size{Width: 300},
		HSpan:   1,
		VSpan:   5,
		VAlign:  align.Middle,
	})
	panel.AddChild(lblAppKey)
	panel.AddChild(inpAppKey)
	panel.AddChild(lblAppSecret)
	panel.AddChild(inpAppSecret)
	panel.AddChild(lblAuthCode)
	panel.AddChild(inpAuthCode)
	panel.Pack()
	return panel
}

func save() {
	_settings.WindowRect = mainWindow.FrameRect()
	_settings.AppAuth.AppKey = inpAppKey.Text()
	_settings.AppAuth.AppSecret = inpAppSecret.Text()
	token, err := api.RequestRefreshToken(_settings.AppAuth, inpAuthCode.Text())
	if err == nil {
		_settings.RefreshToken = token
		api.SetConnectionData(_settings.AppAuth, token)
	}
	fmt.Println(token)
	saveSettings()
}

func getAuthorizationCode() {
	var auth api.AppAuthType
	auth.AppKey = inpAppKey.Text()
	auth.AppSecret = inpAppSecret.Text()
	inpAuthCode.SetText("")
	_settings.RefreshToken = ""
	err := api.AuthorizeApp(auth)
	if err == nil {
		inpAuthCode.SetEnabled(true)
		authorizeSucceeded = true
	}
}

func inpModifiedCallback(_, _ *unison.FieldState) {
	okButton.SetEnabled(checkOk())
	authButton.SetEnabled(checkAuth())
}

func checkOk() bool {
	return inpAppSecret.Text() != "" && inpAppKey.Text() != "" && inpAuthCode.Text() != "" && authorizeSucceeded
}

func checkAuth() bool {
	return inpAppSecret.Text() != "" && inpAppKey.Text() != ""
}
