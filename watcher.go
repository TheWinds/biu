package main

import (
	"log"
	"os"
	"path/filepath"

	"strings"

	"io/ioutil"

	"fmt"

	"github.com/howeyc/fsnotify"
)

//处理文件夹
func dealFolder(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	if event.IsCreate() {
		log.Println("新增文件夹:", event.Name)
		folders[event.Name] = true
		watcher.Watch(event.Name)
	}
	if event.IsDelete() {
		log.Println("删除文件夹:", event.Name)
		watcher.RemoveWatch(event.Name)
	}
	if event.IsRename() {
		log.Println("重命名文件夹:", event.Name)

		//如果是windows重启APP
		if osType == "windows" {
			fmt.Println("重启")
			reStart <- true
		}
		//log.Println(watcher.RemoveWatch(event.Name))
		// os.Mkdir(event.Name, 0777)
		//log.Println("err:", watcher.Remove(event.Name))
	}
}

//处理文件
func dealFile(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	if event.IsModify() {
		log.Println("修改文件:", event.Name)
	}
	if event.IsDelete() {
		log.Println("删除文件:", event.Name)
	}
	if event.IsRename() {
		log.Println("重命名文件:", event.Name)
	}
	if event.IsCreate() {
		log.Println("新增文件:", event.Name)
		src, err := ioutil.ReadFile(event.Name)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(src))
	}
}

//被观察文件的拓展名
var watchExts = []string{
	".js",
	".html",
	".css",
	".gif",
	".png",
	".jpg",
}

// shouldWatchExt 检查是不是应该观察的拓展名
func shouldWatchExt(eventName string) bool {
	for _, ext := range watchExts {
		if strings.HasSuffix(strings.ToLower(eventName), ext) {
			return true
		}
	}
	return false
}

//事件处理器
func eventHandler(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	//检查是否是文件夹
	log.Println(event.String())
	if isFolder(event, workDir) {
		dealFolder(watcher, event)
		return
	}
	//处理文件
	if !shouldWatchExt(event.Name) {
		return
	}
	dealFile(watcher, event)
}

// isFolder判断是否为文件夹
func isFolder(event fsnotify.FileEvent, workDir string) bool {
	if event.IsDelete() {
		_, exist := folders[event.Name]
		if exist {
			//判断文件夹时删除文件
			delete(folders, event.Name)
		}
		return exist
	}
	if event.IsRename() {
		_, exist := folders[event.Name]
		if exist {
			//判断文件夹时删除文件
			delete(folders, event.Name)
		}
		fmt.Println("......", exist)
		return exist
	}
	fi, err := os.Stat(workDir + event.Name)
	if err != nil {

		log.Println(err)
		return false
	}

	return fi.IsDir()
}

//readDirAndFile 获取所有符合规则的目录和文件
func readDirAndFile(workDir string) (files, paths []string, err error) {
	files = make([]string, 0, 30)
	paths = make([]string, 0, 30)
	pathMap := make(map[string]bool)                                                      //忽略后缀匹配的大小写
	err = filepath.Walk(workDir, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// 处理目录
			// 获取 .开头的目录
			if strings.HasPrefix(getRelRootFilePath(filename, workDir), ".") {
				return nil
			}
			pathMap[filename] = true
			return nil
		}
		//忽略不该监控的后缀名
		if shouldWatchExt(fi.Name()) {
			files = append(files, getRelRootFilePath(filename, workDir))
		}
		return nil
	})
	//转换为相对工作路径的路径
	for path := range pathMap {
		paths = append(paths, getRelRootFilePath(path, workDir))
	}
	//将工作路径根目录加入路径
	paths = append(paths, "./")
	return files, paths, err
}

//getRelRootFilePath 获取相对路径
func getRelRootFilePath(path, rootPath string) string {
	if path == rootPath {
		return "."
	}
	return strings.Replace(path, rootPath+"/", "", -1)
}
