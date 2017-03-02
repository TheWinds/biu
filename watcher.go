package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"io/ioutil"

	"fmt"

	"github.com/howeyc/fsnotify"
	"github.com/thewinds/biu/filerefmap"
	"github.com/thewinds/biu/setting"
)

// MonitoredDirs 被监视的文件夹
type MonitoredDirs map[string]bool

// Add 添加文件夹到监视列表
func (monitoredDirs MonitoredDirs) Add(dirName string) {
	monitoredDirs[dirName] = true
}

// Remove 添加文件夹到监视列表
func (monitoredDirs MonitoredDirs) Remove(dirName string) {
	delete(monitoredDirs, dirName)
}

// Has 文件夹是否在监视列表中
func (monitoredDirs MonitoredDirs) Has(dirName string) bool {
	_, has := monitoredDirs[dirName]
	return has
}

var monitoredDirs = make(MonitoredDirs)
var fileRefMap *filerefmap.FileRefMap

// StartWatch 开始监听文件
func StartWatch() {
	//初始化文件引用关系图
	fileRefMap := new(filerefmap.FileRefMap)
	//扫描文件
	files, paths, _ := scanDirAndFile()

	fmt.Println(files)
	fmt.Println(paths)

	//初始化监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	//开始监听
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
		// folders[path] = true
	}
	for _, file := range files {
		_, path, filetype := getFileInfo(file)
		fileRefMap.AddFile(filerefmap.FileNode{Path: path, Type: filetype})
	}
	select {}
}

//处理文件夹
func dealDir(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	if event.IsCreate() {
		log.Println("新增文件夹:", event.Name)
		//folders[event.Name] = true
		watcher.Watch(event.Name)
	}
	if event.IsDelete() {
		log.Println("删除文件夹:", event.Name)
		watcher.RemoveWatch(event.Name)
	}
	if event.IsRename() {
		log.Println("重命名文件夹:", event.Name)

		//如果是windows重启APP
		if setting.OS == "windows" {
			fmt.Println("重启")
			// reStart <- true
		}
		//log.Println(watcher.RemoveWatch(event.Name))
		// os.Mkdir(event.Name, 0777)
		//log.Println("err:", watcher.Remove(event.Name))
	}
}

//处理文件
func dealFile(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	//忽略不关注的文件
	if !setting.ShouldWatchFile(event.Name) {
		return
	}
	if event.IsModify() {
		log.Println("修改文件:", formatName(event.Name))
		src, err := ioutil.ReadFile(formatName(event.Name))
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(src))
		return
	}
	if event.IsDelete() {
		log.Println("删除文件:", event.Name)
	}
	if event.IsRename() {
		log.Println("重命名文件:", event.Name)
		return
	}
	if event.IsCreate() {
		log.Println("新增文件:", event.Name)
		src, err := ioutil.ReadFile(event.Name)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println(string(src))
		return
	}
}

//事件处理器
func eventHandler(watcher *fsnotify.Watcher, event fsnotify.FileEvent) {
	//检查是否是文件夹
	log.Println(event.String())
	if isDir(event) {
		dealDir(watcher, event)
		return
	}
	dealFile(watcher, event)
}

// isFolder判断是否为文件夹
func isDir(event fsnotify.FileEvent) bool {
	fileName := formatName(event.Name)
	//被监控过因此是文件夹
	if event.IsDelete() || event.IsRename() {
		return monitoredDirs.Has(fileName)
	}
	//检查存在的文件是否为文件夹
	fi, err := os.Stat(setting.WorkDir + "/" + fileName)
	if err != nil {
		log.Println(err)
		return false
	}
	return fi.IsDir()
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
			// log.Println("文件夹:")
			// log.Println(getRelPath(filename))

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

//formatName
func formatName(filename string) string {
	if strings.HasPrefix(filename, "./") {
		return filename[2:]
	}
	return filename
}
