package main

import (
	"fmt"
	"net/http"

	"log"

	"flag"

	"strconv"

	"os"

	"github.com/fatih/color"
	"github.com/thewinds/biu/setting"
)

func main() {
	initFlag()
	startServ()
	StartWatch()

}
func startServ() {
	wshander := InitNotifyServ()
	// 通知页面刷新的websocket服务
	http.Handle(setting.WSServPath, wshander)
	// 文件服务器
	http.Handle("/", InjectFileServer(http.Dir(""), InjectScriptFunc))
	// 获取要注入的js代码的handler
	http.HandleFunc(setting.InjectScriptPath, InjectScriptHandler)
	color.Green("[Biu] 启动http服务 localhost:" + setting.Port)
	go func() {
		err := http.ListenAndServe(":"+setting.Port, nil)
		if err != nil {
			color.Red("启动http服务器失败")
			os.Exit(1)
		}
	}()
}
func initFlag() {
	port := flag.String("p", "8080", "指定运行的端口")
	help := flag.Bool("help", false, "查看帮助")
	flag.Parse()
	if *help == true {
		color.Red("\nbiu 实时刷新工具 ❤\n")
		fmt.Printf("\n使用帮助\n")
		fmt.Println("biu \t\t运行http服务器在默认端口8080并实时刷新")
		fmt.Println("biu -p=端口号\t运行http服务器在指定端口并实时刷新")
		fmt.Println("biu -help\t查看帮助")
		fmt.Println()
		color.Cyan("Powered by thewinds")
		os.Exit(0)
	}
	if _, err := strconv.Atoi(*port); err != nil {
		log.Fatal("端口不正确,请检查")
	}
	setting.Port = *port
}
