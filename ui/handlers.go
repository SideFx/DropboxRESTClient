// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// Event handlers
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/api"
	"Dropbox_REST_Client/assets"
	"Dropbox_REST_Client/dialogs"
	"Dropbox_REST_Client/models"
	"fmt"
	"github.com/richardwilkes/unison"
	"os"
)

func aboutUser() {
	userinfo, err := api.GetCurrentUser()
	if err == nil {
		dialogs.AboutUserDialog(userinfo)
	}
}

func refresh() {
	models.DropboxRefreshData()
}

func newFolder(isRoot bool) {
	folderName := dialogs.DialogToQueryFolderName()
	if folderName == "" {
		return
	}
	models.DropboxCreateFolder(isRoot, folderName)
}

func deleteItem() {
	models.DropboxDeleteFileItems()
}

func uploadItems() {
	var allItems []*api.FolderStructureType
	homeDir, _ := os.UserHomeDir()
	dialog := unison.NewOpenDialog()
	dialog.SetInitialDirectory(homeDir)
	dialog.SetCanChooseFiles(true)
	dialog.SetAllowsMultipleSelection(true)
	dialog.SetCanChooseDirectories(true)
	dialog.SetCanChooseFiles(true)
	dialog.SetResolvesAliases(false)
	if dialog.RunModal() {
		for _, p := range dialog.Paths() {
			stat, err := os.Stat(p)
			if err != nil {
				dialogs.DialogToDisplaySystemError(assets.ErrorReadError, err)
				fmt.Println(err)
				return
			}
			isFolder := stat.IsDir()
			if isFolder {
				tmp, err := api.ExplodeFolder(p)
				if err != nil {
					dialogs.DialogToDisplaySystemError(assets.ErrorReadError, err)
					return
				}
				allItems = append(allItems, tmp...)
			} else {
				allItems = append(allItems, &api.FolderStructureType{Path: p, IsFolder: isFolder})
			}
		}
	}
	for _, item := range allItems {
		fmt.Println(item.Path)
	}
}

func downloadItems() {

}
