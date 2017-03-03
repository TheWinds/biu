package main

import (
	"fmt"
	"net/http"
	"strings"

	"log"

	"flag"

	"strconv"

	"github.com/thewinds/biu/setting"
	"gopkg.in/macaron.v1"
)

func main() {
	initFlag()
	initServ()
	StartWatch()

}
func initServ() {
	wshander := InitNotifyServ()
	m := macaron.New()
	m.Use(func(ctx *macaron.Context) {
		log.Println(ctx.Req.URL.String())
		if strings.HasSuffix(ctx.Req.URL.String(), ".html") || strings.HasSuffix(ctx.Req.URL.String(), "/") {
			ctx.Write([]byte(`<script src="` + setting.InjectScriptPath + `"></script>`))
		}
	})
	m.Use(macaron.Static(""))
	m.Get(setting.InjectScriptPath, func() string {
		return GetInjectScript()
	})
	log.Println(setting.WSServPath)
	http.Handle(setting.WSServPath, wshander)
	http.Handle("/", m)
	log.SetPrefix("[Biu] ")

	log.Println("启动http服务 localhost:" + setting.Port)
	go http.ListenAndServe(":"+setting.Port, nil)
}
func initFlag() {
	port := flag.String("p", "8080", "指定运行的端口")
	help := flag.Bool("help", false, "查看帮助")
	flag.Parse()
	if *help == true {
		fmt.Println("biu 实时刷新工具\n")
		fmt.Println("使用帮助\n")
		fmt.Println("biu \t\t运行http服务器在默认端口8080并实时刷新")
		fmt.Println("biu -p=端口号\t运行http服务器在指定端口并实时刷新")
		fmt.Println("biu -help\t查看帮助")
		fmt.Println()
		fmt.Println("Powered by thewinds")
	}
	if _, err := strconv.Atoi(*port); err != nil {
		log.Fatal("端口不正确,请检查")
	}
	setting.Port = *port
}
