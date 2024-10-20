// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// OS utilities
// ---------------------------------------------------------------------------------------------------------------------

package api

import (
	"os"
	"path/filepath"
)

type FolderStructureType struct {
	Path     string
	IsFolder bool
}

func ExplodeFolder(folder string) ([]*FolderStructureType, error) {
	var folderStructure []*FolderStructureType
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			folderStructure = append(folderStructure, &FolderStructureType{Path: path, IsFolder: info.IsDir()})
			return nil
		})
	if err != nil {
		return nil, err
	}
	return folderStructure, err
}
