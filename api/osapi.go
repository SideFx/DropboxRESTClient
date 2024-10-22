// ---------------------------------------------------------------------------------------------------------------------
// (w) 2024 by Jan Buchholz
// OS utilities
// ---------------------------------------------------------------------------------------------------------------------

package api

import (
	"os"
	"path/filepath"
	"strings"
)

type FileSysStructureType struct {
	OSPath   string
	DbxPath  string
	IsFolder bool
	Size     int64
}

func ExplodeFolder(folder string) ([]*FileSysStructureType, error) {
	var folderStructure []*FileSysStructureType
	basepath := filepath.Base(folder) + string(os.PathSeparator)
	prefixpath := strings.TrimSuffix(folder, basepath)
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			_, file := filepath.Split(path)
			if (file == "") || (file[0] != '.') {
				shortpath := DbxPathSeparator + strings.TrimPrefix(path, prefixpath)
				shortpath = strings.ReplaceAll(shortpath, string(os.PathSeparator), DbxPathSeparator) // Windows
				folderStructure = append(folderStructure,
					&FileSysStructureType{OSPath: path, DbxPath: shortpath, IsFolder: info.IsDir(), Size: info.Size()})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return folderStructure, err
}

func PrepareDbxUpload(selection []string) ([]*FileSysStructureType, []*FileSysStructureType, error) {
	var folders []*FileSysStructureType
	var files []*FileSysStructureType
	var err error
	for _, s := range selection {
		stat, err := os.Stat(s)
		if err != nil {
			return nil, nil, err
		}
		if !stat.IsDir() {
			_, file := filepath.Split(s)
			if file[0] != '.' {
				files = append(files, &FileSysStructureType{
					OSPath:   s,
					DbxPath:  DbxPathSeparator + file,
					IsFolder: true,
					Size:     stat.Size()},
				)
			}
		} else {
			folderStructures, err := ExplodeFolder(s)
			if err != nil {
				return nil, nil, err
			}
			for _, fs := range folderStructures {
				if fs.IsFolder {
					folders = append(folders, fs)
				} else {
					files = append(files, fs)
				}
			}
		}
	}
	return folders, files, err
}
