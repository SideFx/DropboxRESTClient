// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// Dialogs, using Unison library (c) Richard A. Wilkes
// https://github.com/richardwilkes/unison
// ---------------------------------------------------------------------------------------------------------------------

package dialogs

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"errors"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"strings"
)

const (
	imageWidth  float32 = 150
	imageHeight float32 = 150
)

func AboutDialog(item unison.MenuItem) {
	dialog, err := unison.NewDialog(nil, nil, newAboutPanel(),
		[]*unison.DialogButtonInfo{unison.NewOKButtonInfo()},
		unison.NotResizableWindowOption())
	if err == nil {
		wnd := dialog.Window()
		wnd.SetTitle(item.Title())
		//if len(titleIcons) > 0 {
		//	wnd.SetTitleIcons(titleIcons)
		//}
		okButton := dialog.Button(unison.ModalResponseOK)
		okButton.ClickCallback = func() {
			dialog.StopModal(unison.ModalResponseOK)
		}
		dialog.RunModal()
	}
}

func newAboutPanel() *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	//breakTextIntoLabels(panel, assets.TxtAboutEmbyExplorer, unison.LabelFont.Face().Font(10), false, true)
	//breakTextIntoLabels(panel, assets.TxtAboutUnison, unison.LabelFont.Face().Font(10), false, true)
	//breakTextIntoLabels(panel, assets.TxtAboutExcelize, unison.LabelFont.Face().Font(10), false, true)
	panel.SetLayoutData(&unison.FlexLayoutData{
		MinSize: unison.Size{Width: 500},
		HSpan:   1,
		VSpan:   1,
		VAlign:  align.Middle,
	})
	return panel
}

// taken (and slightly modifield) from Unison dialog.go
func breakTextIntoLabels(panel *unison.Panel, text string, font unison.Font, addSpaceAbove bool, center bool) {
	if text != "" {
		returns := 0
		for {
			if i := strings.Index(text, "\n"); i != -1 {
				if i == 0 {
					returns++
					text = text[1:]
				} else {
					part := text[:i]
					l := unison.NewLabel()
					l.Font = font
					l.SetTitle(part)
					if center {
						l.SetLayoutData(&unison.FlexLayoutData{
							HSpan:  1,
							VSpan:  1,
							HAlign: align.Middle,
							VAlign: align.Middle,
							HGrab:  true,
						})
					}
					if returns > 1 || addSpaceAbove {
						addSpaceAbove = false
						l.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdHSpacing}))
					}
					panel.AddChild(l)
					text = text[i+1:]
					returns = 1
				}
			} else {
				if text != "" {
					l := unison.NewLabel()
					l.Font = font
					l.SetTitle(text)
					if center {
						l.SetLayoutData(&unison.FlexLayoutData{
							HSpan:  1,
							VSpan:  1,
							HAlign: align.Middle,
							VAlign: align.Middle,
							HGrab:  true,
						})
					}
					if returns > 1 || addSpaceAbove {
						l.SetBorder(unison.NewEmptyBorder(unison.Insets{Top: unison.StdHSpacing}))
					}
					panel.AddChild(l)
				}
				break
			}
		}
	}
}

func DialogToDisplaySystemError(primary string, detail error) {
	var msg string
	var err errs.StackError
	if errors.As(detail, &err) {
		errs.Log(detail)
		msg = err.Message()
	} else {
		msg = detail.Error()
	}
	panel := unison.NewMessagePanel(primary, msg)
	if dialog, err := unison.NewDialog(unison.DefaultDialogTheme.ErrorIcon, unison.DefaultDialogTheme.ErrorIconInk, panel,
		[]*unison.DialogButtonInfo{unison.NewOKButtonInfo()}, unison.NotResizableWindowOption()); err != nil {
		errs.Log(err)
	} else {
		wnd := dialog.Window()
		wnd.SetTitle(assets.CapError)
		//if len(titleIcons) > 0 {
		//	wnd.SetTitleIcons(titleIcons)
		//}
		dialog.RunModal()
	}
}

func AboutUserDialog(userinfo *api.UserInfoType) {
	var image *unison.Image
	var frame unison.Rect
	var imagePanel *unison.Panel
	var cols int = 1
	if userinfo.ProfilePhotoUrl != "" {
		rawdata, _ := api.CurrentUserGetPicture(userinfo.ProfilePhotoUrl)
		if rawdata != nil {
			image, _ = unison.NewImageFromBytes(rawdata, 1)
		}
	}
	if image != nil {
		imagePanel = newImagePanel(image)
		cols = 2
	}
	wnd, err := unison.NewWindow(assets.CapAboutUser, unison.NotResizableWindowOption())
	if err != nil {
		panic(err)
	}
	if focused := unison.ActiveWindow(); focused != nil {
		frame = focused.FrameRect()
	} else {
		frame = unison.PrimaryDisplay().Usable
	}
	//if len(titleIcons) > 0 {
	//	wnd.SetTitleIcons(titleIcons)
	//}
	content := wnd.Content()
	content.SetLayout(&unison.FlexLayout{
		Columns:  cols,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	content.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(15)))
	if imagePanel != nil {
		content.AddChild(imagePanel)
	}
	content.AddChild(newDetailsPanel(userinfo))
	buttonPanel := unison.NewPanel()
	buttonPanel.SetLayout(&unison.FlexLayout{
		Columns:      2,
		EqualColumns: true,
	})
	buttonPanel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		VSpan:  1,
		HAlign: align.Middle,
		VAlign: align.Middle,
	})
	okButton := unison.NewButton()
	okButton.SetTitle(assets.CapClose)
	okButton.ClickCallback = func() {
		wnd.StopModal(0)
		wnd.Dispose()
	}
	buttonPanel.AddChild(okButton)
	content.AddChild(buttonPanel)
	wnd.Pack()
	wndFrame := wnd.FrameRect()
	frame.Y += (frame.Height - wndFrame.Height) / 3
	frame.Height = wndFrame.Height
	frame.X += (frame.Width - wndFrame.Width) / 2
	frame.Width = wndFrame.Width
	wnd.SetFrameRect(frame.Align())
	wnd.RunModal()
}

func newDetailsPanel(userinfo *api.UserInfoType) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 10,
		VSpacing: unison.StdVSpacing,
	})
	lblAccountId := unison.NewLabel()
	lblAccountId.Font = unison.LabelFont
	lblAccountId.SetTitle(assets.CapAccountId)
	panel.AddChild(lblAccountId)
	lblAccountIdDisp := unison.NewLabel()
	lblAccountIdDisp.Font = unison.LabelFont
	lblAccountIdDisp.SetTitle(userinfo.AccountId)
	panel.AddChild(lblAccountIdDisp)
	lblAccountType := unison.NewLabel()
	lblAccountType.Font = unison.LabelFont
	lblAccountType.SetTitle(assets.CapAccountType)
	panel.AddChild(lblAccountType)
	lblAccountTypeDisp := unison.NewLabel()
	lblAccountTypeDisp.Font = unison.LabelFont
	lblAccountTypeDisp.SetTitle(userinfo.AccountType.Tag)
	panel.AddChild(lblAccountTypeDisp)
	lblCountry := unison.NewLabel()
	lblCountry.Font = unison.LabelFont
	lblCountry.SetTitle(assets.CapCountry)
	panel.AddChild(lblCountry)
	lblCountryDisp := unison.NewLabel()
	lblCountryDisp.Font = unison.LabelFont
	lblCountryDisp.SetTitle(userinfo.Country)
	panel.AddChild(lblCountryDisp)
	lblEmail := unison.NewLabel()
	lblEmail.Font = unison.LabelFont
	lblEmail.SetTitle(assets.CapEmail)
	panel.AddChild(lblEmail)
	lblEmailDisp := unison.NewLabel()
	lblEmailDisp.Font = unison.LabelFont
	lblEmailDisp.SetTitle(userinfo.Email)
	panel.AddChild(lblEmailDisp)
	lblName := unison.NewLabel()
	lblName.Font = unison.LabelFont
	lblName.SetTitle(assets.CapDisplayName)
	panel.AddChild(lblName)
	lblNameDisp := unison.NewLabel()
	lblNameDisp.Font = unison.LabelFont
	lblNameDisp.SetTitle(userinfo.Name.DisplayName)
	panel.AddChild(lblNameDisp)
	lblNamespace := unison.NewLabel()
	lblNamespace.Font = unison.LabelFont
	lblNamespace.SetTitle(assets.CapNamespace)
	panel.AddChild(lblNamespace)
	lblNamespaceDisp := unison.NewLabel()
	lblNamespaceDisp.Font = unison.LabelFont
	lblNamespaceDisp.SetTitle(userinfo.RootInfo.HomeNamespaceId)
	panel.AddChild(lblNamespaceDisp)
	panel.Pack()
	return panel
}

func newImagePanel(image *unison.Image) *unison.Panel {
	panel := unison.NewPanel()
	imgPanel := unison.NewLabel()
	imgPanel.Drawable = image
	imgPanel.SetBorder(unison.NewEmptyBorder(unison.NewUniformInsets(1)))
	_, prefSize, _ := imgPanel.Sizes(unison.Size{})
	prefSize.Width = imageWidth
	prefSize.Height = imageHeight
	imgPanel.SetFrameRect(unison.Rect{Size: prefSize})
	panel.AddChild(imgPanel.AsPanel())
	panel.SetLayoutData(&unison.FlexLayoutData{
		SizeHint: prefSize,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
		HGrab:    true,
		VGrab:    true,
	})
	return panel
}

func DialogToDisplayErrorMessage(primary string, detail string) {
	panel := unison.NewMessagePanel(primary, detail)
	if dialog, err := unison.NewDialog(unison.DefaultDialogTheme.ErrorIcon, unison.DefaultDialogTheme.ErrorIconInk, panel,
		[]*unison.DialogButtonInfo{unison.NewOKButtonInfo()}, unison.NotResizableWindowOption()); err != nil {
		errs.Log(err)
	} else {
		wnd := dialog.Window()
		wnd.SetTitle(assets.CapError)
		//if len(titleIcons) > 0 {
		//	wnd.SetTitleIcons(titleIcons)
		//}
		dialog.RunModal()
	}
}

const inpTextSizeMax = 250

func DialogToQueryFolderName() string {
	var dialog *unison.Dialog
	var err error
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 10,
		VSpacing: unison.StdVSpacing,
	})
	lblName := unison.NewLabel()
	lblName.Font = unison.LabelFont
	lblName.SetTitle(assets.CapFolderName)
	inpName := unison.NewField()
	inpName.Font = unison.FieldFont
	inpName.MinimumTextWidth = inpTextSizeMax
	inpName.ModifiedCallback = func(before, after *unison.FieldState) {
		dialog.Button(unison.ModalResponseOK).SetEnabled(after.Text != "" && api.CheckNameIsValid(after.Text))
	}
	panel.AddChild(lblName)
	panel.AddChild(inpName)
	if dialog, err = unison.NewDialog(nil, nil, panel,
		[]*unison.DialogButtonInfo{unison.NewCancelButtonInfo(), unison.NewOKButtonInfo()},
		unison.NotResizableWindowOption()); err != nil {
		errs.Log(err)
	} else {
		wnd := dialog.Window()
		wnd.SetTitle(assets.CapCreateFolder)
		//if len(titleIcons) > 0 {
		//	wnd.SetTitleIcons(titleIcons)
		//}
		dialog.Button(unison.ModalResponseOK).SetEnabled(false)
		dialog.Button(unison.ModalResponseCancel).ClickCallback = func() {
			inpName.SetText("")
			dialog.StopModal(unison.ModalResponseCancel)
		}
		dialog.RunModal()
	}
	return inpName.Text()
}
