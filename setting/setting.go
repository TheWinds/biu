package setting

import (
	"log"
	"os"
	"runtime"
	"strings"
)

func init() {
	var err error
	WorkDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	OS = runtime.GOOS
}

// WorkDir 工作路径
var WorkDir string

// OS 系统类型
var OS string

//被观察文件的拓展名
var watchExts = []string{
	".js",
	".html",
	".css",
	".gif",
	".png",
	".jpg",
}

// ShouldWatchFile 检查是不是应该观察的拓展名
func ShouldWatchFile(fileName string) bool {
	for _, ext := range watchExts {
		if strings.HasSuffix(strings.ToLower(fileName), ext) {
			return true
		}
	}
	return false
}
