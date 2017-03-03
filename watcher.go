package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"io/ioutil"

	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/thewinds/biu/filerefmap"
	"github.com/thewinds/biu/reffinder"
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
	fileRefMap = new(filerefmap.FileRefMap)
	//扫描文件
	files, paths, _ := scanDirAndFile()
	//初始化监听器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	//开始监听
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				eventHandler(watcher, event)
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
	color.Green("[Biu] 开始监听代码改动")
	color.Red("[Biu] 保存文件后相关文件会自动刷新 ❤")
	for _, path := range paths {
		err = watcher.Add(path)

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
	for _, file := range files {
		updateFileRef(file)
	}
	select {}
}

func updateFileRef(fileName string) {
	fileName = formatName(fileName)
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

// isOp 判断操作是否为指定操作
func isOp(event fsnotify.Event, op fsnotify.Op) bool {
	return event.Op&op == op

}

// EventTimeLine 事件时间线
type EventTimeLine struct {
	EventInfo   string
	HappendTime int64
}

var eventTimeLine = new(EventTimeLine)

// 是否应该忽略重复的事件
func shouldIngoreEvent(event fsnotify.Event) bool {
	eventInfo := event.String()
	eventTime := time.Now().UnixNano()

	if eventTimeLine.EventInfo == eventInfo {
		if eventTime-eventTimeLine.HappendTime > setting.ScanFilePeriod {
			eventTimeLine.EventInfo = eventInfo
			eventTimeLine.HappendTime = eventTime
			return false
		}
		return true
	}
	eventTimeLine.EventInfo = eventInfo
	eventTimeLine.HappendTime = eventTime
	return false
}

//事件处理器
func eventHandler(watcher *fsnotify.Watcher, event fsnotify.Event) {
	if shouldIngoreEvent(event) {
		return
	}
	//检查是否是文件夹

	if isDir(event) {
		dealDir(watcher, event)
		return
	}
	dealFile(watcher, event)
}

//处理文件夹
func dealDir(watcher *fsnotify.Watcher, event fsnotify.Event) {
	dirName := formatName(event.Name)
	if isOp(event, fsnotify.Create) {
		// log.Println("新增文件夹:", dirName)
		monitoredDirs.Add(dirName)
		watcher.Add(dirName)
	}
	if isOp(event, fsnotify.Remove) {
		// log.Println("删除文件夹:", dirName)
		monitoredDirs.Remove(dirName)
		fileRefMap.RemoveDirFile(dirName)
		watcher.Remove(dirName)
	}
	if isOp(event, fsnotify.Rename) {
		// log.Println("重命名文件夹:", dirName)
		monitoredDirs.Remove(dirName)
		fileRefMap.RemoveDirFile(dirName)
	}
}

//处理文件
func dealFile(watcher *fsnotify.Watcher, event fsnotify.Event) {
	fileName := formatName(event.Name)
	//忽略不关注的文件
	if !setting.ShouldWatchFile(fileName) {
		return
	}
	// 判断事件类型
	if isOp(event, fsnotify.Write) {
		updateFileRef(fileName)
		//如果是html文件的改动就直接通知
		if strings.HasSuffix(fileName, ".html") {
			NotifyRefresh(fileName)
			return
		}
		//否则检测到根html节点然后通知所有相应文件刷新
		files := fileRefMap.FindRoots(fileName)
		NotifyMultiRefresh(files)

		return
	}
	if isOp(event, fsnotify.Remove) {
		// log.Println("删除文件:", fileName)

		files := fileRefMap.FindRoots(fileName)
		NotifyMultiRefresh(files)
		fileRefMap.RemoveFile(fileName)
	}
	if isOp(event, fsnotify.Rename) {
		// log.Println("重命名文件:", fileName)
		fileRefMap.RemoveFile(fileName)
		return
	}
	if isOp(event, fsnotify.Create) {
		// log.Println("新增文件:", fileName)
		_, path, filetype := getFileInfo(fileName)
		fileRefMap.AddFile(filerefmap.FileNode{Path: path, Type: filetype})
		updateFileRef(fileName)
		return
	}
}

// isFolder判断是否为文件夹
func isDir(event fsnotify.Event) bool {
	fileName := formatName(event.Name)
	//被监控过因此是文件夹
	if isOp(event, fsnotify.Remove) || isOp(event, fsnotify.Rename) {
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
	dirs = append(dirs, "")
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
