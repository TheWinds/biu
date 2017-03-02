package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/thewinds/biu/filerefmap"
)

var (
	//工作路径
	workDir string
	//被监视的文件夹
	folders map[string]bool
	//文件引用图
	fileRefMap *filerefmap.FileRefMap
	//系统类型
	osType string
	//是否重启
	reStart chan bool
)

func main() {
	//初始化
	osType = runtime.GOOS
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	// filepath
	// if !strings.HasSuffix(workDir, "\\") {
	// 	workDir = workDir + "\\"
	// }
	//执行观察任务
	for {
		folders = make(map[string]bool)

		fileRefMap = new(filerefmap.FileRefMap)
		fmt.Println(workDir)
		files, paths, _ := readDirAndFile(workDir)
		fmt.Println(files)
		fmt.Println(paths)
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()
		go func() {
			for {
				select {
				case event := <-watcher.Event:
					eventHandler(watcher, *event)
				case err := <-watcher.Error:
					log.Println("error:", err)
				}
			}
		}()
		log.Println("共有", len(paths), "个目录")
		for _, path := range paths {
			log.Println("|" + path + "|")
			err = watcher.Watch(path)

			if err != nil {
				log.Fatal(err, "“"+path+"”")
			}
			//加入文件夹列表
			folders[path] = true
		}
		for _, file := range files {
			_, path, filetype := getFileInfo(file)
			fileRefMap.AddFile(filerefmap.FileNode{Path: path, Type: filetype})
		}

		shouldReStart := <-reStart
		if !shouldReStart {
			break
		}
	}

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
