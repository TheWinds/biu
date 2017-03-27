package setting

import (
	"encoding/base64"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

func init() {
	var err error
	WorkDir, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	OS = runtime.GOOS
	ScanFilePeriod = int64(time.Millisecond * 10)
	Port = "8080"
	WSServPath = "/wsserv"
	WSConnKey = base64.URLEncoding.EncodeToString([]byte(time.Now().String()))
	InjectScriptPath = "/" + base64.URLEncoding.EncodeToString([]byte(time.Now().AddDate(3, 4, 5).String())) + ".js"
}

var (
	// WorkDir 工作路径
	WorkDir string
	// OS 系统类型
	OS string
	// ScanFilePeriod 扫描文件的周期
	ScanFilePeriod int64
	// WSConnKey websocket连接秘钥
	WSConnKey string
	// WSServPath websocket服务地址
	WSServPath string
	// Port websocket服务端口
	Port string
	// InjectScriptPath 被注入的脚本地址
	InjectScriptPath string

	// Logo
	Logo = `
______   _         
| ___ \ (_)        
| |_/ /  _   _   _ 
| ___ \ | | | | | |
| |_/ / | | | |_| |
\____/  |_|  \__,_|
                    v1.1				   
`
)

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
