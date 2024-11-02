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
	FileName string
	IsFolder bool
	Size     int64
}

func ExplodeFolder(folder string) ([]*FileSysStructureType, error) {
	var folderStructure []*FileSysStructureType
	basepath := filepath.Base(folder) + string(os.PathSeparator)
	prefixpath := strings.TrimSuffix(folder, basepath)
	err := filepath.Walk(folder,
		func(path string, info os.FileInfo, err error) error {
			var _path, _file string
			if err != nil {
				return err
			}
			_path, _file = filepath.Split(path)
			if info.IsDir() {
				_path = _path + _file
			}
			// omit dot files and folders
			if (_file == "") || (_file[0] != '.') {
				shortpath := DbxPathSeparator + strings.TrimPrefix(_path, prefixpath)
				shortpath = strings.ReplaceAll(shortpath, string(os.PathSeparator), DbxPathSeparator) // Windows
				folderStructure = append(folderStructure,
					&FileSysStructureType{
						OSPath:   path,
						DbxPath:  shortpath,
						FileName: _file, //replaceInvalidChars(file),
						IsFolder: info.IsDir(),
						Size:     info.Size(),
					})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}
	return folderStructure, nil
}

func ListLocalFileStructure(selection []string) ([]*FileSysStructureType, []*FileSysStructureType, error) {
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
					DbxPath:  DbxPathSeparator,
					FileName: file, //replaceInvalidChars(file),
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

func replaceInvalidChars(filename string) string {
	name := filename
	if name != "" {
		for _, char := range DbxInvalidCharacters {
			name = strings.ReplaceAll(name, string(char), DbxReplaceBySubst)
		}
	}
	return name
}
