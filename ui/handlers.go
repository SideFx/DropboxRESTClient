// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// Event handlers
// ---------------------------------------------------------------------------------------------------------------------

package ui

import (
	"Dropbox_REST_Client/api"
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

func newFolder() {
	folderName := dialogs.DialogToQueryFolderName()
	if folderName == "" {
		return
	}
	models.DropboxCreateFolder(folderName)
}

func deleteItem() {
	models.DropboxDeleteFileItems()
}

func uploadItems() {
	var allFolders, allFiles []*api.FileSysStructureType
	var err error
	homeDir, _ := os.UserHomeDir()
	dialog := unison.NewOpenDialog()
	dialog.SetInitialDirectory(homeDir)
	dialog.SetCanChooseFiles(true)
	dialog.SetAllowsMultipleSelection(true)
	dialog.SetCanChooseDirectories(true)
	dialog.SetCanChooseFiles(true)
	dialog.SetResolvesAliases(false)
	if dialog.RunModal() {
		allFolders, allFiles, err = api.PrepareDbxUpload(dialog.Paths())
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	for _, folder := range allFolders {
		fmt.Println(folder.DbxPath)
	}
	for _, file := range allFiles {
		fmt.Println(file.DbxPath)
	}
}

func downloadItems() {
	filename := "/Users/jan/Downloads/milky-way-nasa.jpg"
	content, _ := os.ReadFile(filename)
	fmt.Println(len(content))
	result := api.ConputeHash(content)
	fmt.Println(result)
}
