// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// Preferences, window rect, authorization settings
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"encoding/json"
	"github.com/richardwilkes/unison"
	"io"
	"os"
	"path/filepath"
)

const preferencesFileName = "org.janbuchholz.dropboxrestclient.json"

type settings struct {
	WindowRect   unison.Rect
	AppAuth      api.AppAuthType
	RefreshToken string
}

var _settings settings

func saveSettings() {
	rect := mainWindow.FrameRect()
	prefs := settings{
		WindowRect:   rect,
		AppAuth:      _settings.AppAuth,
		RefreshToken: _settings.RefreshToken,
	}
	j, err := json.Marshal(prefs)
	if err == nil {
		dir, _ := os.UserConfigDir()
		dir = filepath.Join(dir, assets.AppName)
		_, err := os.Stat(dir)
		if err != nil {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				panic(err)
			}
		}
		fname := filepath.Join(dir, preferencesFileName)
		_ = os.WriteFile(fname, j, 0644)
	}
	if _settings.AppAuth.AppKey != "" && _settings.AppAuth.AppSecret != "" {
		userInfoBtn.SetEnabled(true)
	}
}

func loadSettings() {
	dir, err := os.UserConfigDir()
	dir = filepath.Join(dir, assets.AppName)
	fname := filepath.Join(dir, preferencesFileName)
	j, err := os.Open(fname)
	if err == nil {
		byteValue, _ := io.ReadAll(j)
		_ = j.Close()
		_ = json.Unmarshal(byteValue, &_settings)
		api.SetConnectionData(_settings.AppAuth, _settings.RefreshToken)
	}
}

func IsTokenPresent() bool {
	return _settings.RefreshToken != ""
}
