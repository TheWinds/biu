package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"io/ioutil"

	"github.com/fatih/color"
	"github.com/radovskyb/watcher"
	"github.com/thewinds/biu/filerefmap"
	"github.com/thewinds/biu/reffinder"
	"github.com/thewinds/biu/setting"
)

var fileRefMap *filerefmap.FileRefMap

// StartWatch 开始监听文件
func StartWatch() {
	//初始化文件引用关系图
	fileRefMap = new(filerefmap.FileRefMap)
	//初始化监听器
	fwatcher := watcher.New()
	fwatcher.FilterOps(watcher.Rename, watcher.Move, watcher.Remove, watcher.Create, watcher.Write)
	fwatcher.IgnoreHiddenFiles(true)
	//开始监听
	go func() {
		for {
			select {
			case event := <-fwatcher.Event:
				handlerFileEvent(fwatcher, event)
			case err := <-fwatcher.Error:
				log.Fatal("error:", err)
			case <-fwatcher.Closed:
				return
			}
		}
	}()
	color.Green("[Biu] 开始监听代码改动")
	color.Red("[Biu] 保存文件后相关页面会自动刷新 ❤")
	fwatcher.AddRecursive(".")
	watchfiles := make([]string, 0)
	// 初始化文件和引用关系
	for filepath, fileinfo := range fwatcher.WatchedFiles() {
		//初始化文件
		if !fileinfo.IsDir() && setting.ShouldWatchFile(filepath) {
			filerelpath, err := getRelPath(filepath)
			if err != nil {
				log.Fatal("初始化文件失败", err)
			}
			_, path, filetype := getFileInfo(filerelpath)
			fileRefMap.AddFile(filerefmap.FileNode{Path: path, Type: filetype})
			watchfiles = append(watchfiles, filerelpath)
		}
	}
	for _, filepath := range watchfiles {
		// color.Cyan(filepath)
		NumOfWatcherFiles++
		updateFileRef(filepath)
	}
	updateTerm()
	//开始监听
	if err := fwatcher.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}
	select {}
}

// 更新文件引用关系
func updateFileRef(fileName string) {
	// fileName = formatPath(fileName)
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	reffiles, err := reffinder.FindFileRef(data, fileName, setting.WorkDir)
	if err != nil {
		log.Println(err)
		return
	}
	fileRefMap.UpdateRef(fileName, reffiles)
}

//事件处理器
func handlerFileEvent(fwatcher *watcher.Watcher, event watcher.Event) {
	// log.Println(event)
	// 忽略文件夹
	if event.IsDir() {
		return
	}
	fileName := formatPath(event.Path)
	//忽略不关注的文件
	if !setting.ShouldWatchFile(fileName) {
		return
	}

	switch event.Op {
	case watcher.Create:
		_, path, filetype := getFileInfo(fileName)
		fileRefMap.AddFile(filerefmap.FileNode{Path: path, Type: filetype})
		updateFileRef(fileName)
		NumOfWatcherFiles++
		updateTerm()
	case watcher.Remove:
		files := fileRefMap.FindRoots(fileName)
		NotifyMultiRefresh(files)
		fileRefMap.RemoveFile(fileName)
		NumOfWatcherFiles--
		updateTerm()
	case watcher.Move, watcher.Rename:
		paths := getRenamePath(fileName)
		if paths == nil {
			return
		}
		fileRefMap.ReNameFile(paths[0], paths[1])
		updateFileRef(paths[1])
	case watcher.Write:
		updateFileRef(fileName)
		//如果是html文件的改动就直接通知
		if strings.HasSuffix(fileName, ".html") {
			NotifyRefresh(fileName)
			return
		}
		//否则检测到根html节点然后通知所有相应文件刷新
		files := fileRefMap.FindRoots(fileName)
		NotifyMultiRefresh(files)
	}
}

//scanDirAndFile 扫描工作目录获取所有符合规则的目录和文件
func scanDirAndFile() (files, dirs []string, err error) {
	files = make([]string, 0, 30)
	dirs = make([]string, 0, 30)                                                                  //忽略后缀匹配的大小写
	err = filepath.Walk(setting.WorkDir, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil {
			return err
		}
		if fi.IsDir() {
			// 处理目录
			if dir, err := getRelPath(filename); err != nil {
				return nil
			} else {
				//忽略开头为 . 的文件夹
				if strings.HasPrefix(dir, ".") {
					return nil
				}
				dirs = append(dirs, dir)
			}
			return nil
		}
		//忽略不该监控的后缀名
		if setting.ShouldWatchFile(fi.Name()) {

			if file, err := getRelPath(filename); err == nil {
				files = append(files, file)
				// log.Println("文件:")
				// log.Println(file)
			}
		}
		return nil
	})
	//将工作路径根目录加入路径
	dirs = append(dirs, ".")
	return files, dirs, err
}

//getRelPath	 获取相对工作目录的路径
func getRelPath(path string) (string, error) {
	return filepath.Rel(setting.WorkDir, path)
}

//formatPath
func formatPath(filepath string) string {
	if strings.Contains(filepath, " -> ") {
		return filepath
	}
	filepath, _ = getRelPath(filepath)
	return filepath
}

// getRenamePath 获取更名前和更名后的两个路径
func getRenamePath(filepath string) []string {
	paths := strings.Split(filepath, " -> ")
	if len(paths) != 2 {
		return nil
	}
	for i := 0; i < len(paths); i++ {
		newpath, err := getRelPath(paths[i])
		paths[i] = newpath
		if err != nil {
			return nil
		}
	}
	return paths
}

// 获取文件信息
func getFileInfo(filePath string) (name, path string, filetype filerefmap.FileType) {
	name = filepath.Base(filePath)
	path = filePath
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case "js":
		filetype = filerefmap.JSFile
	case "css":
		filetype = filerefmap.CSSFile
	case "html":
		filetype = filerefmap.HTMLFile
	}
	return
}
