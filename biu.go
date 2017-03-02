package main

import (
	"strings"

	"path/filepath"

	"github.com/thewinds/biu/filerefmap"
)

func main() {
	StartWatch()
}
func getFileInfo(filePath string) (name, path string, filetype filerefmap.FileType) {
	name = filepath.Base(filePath)
	path = filePath
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case "js":
		filetype = filerefmap.JSFile
	case "css":
		filetype = filerefmap.CSSFile
	case "HTML":
		filetype = filerefmap.HTMLFile
	}
	return
}
