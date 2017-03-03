package main

import (
	"net/http"

	"github.com/thewinds/biu/setting"
	"golang.org/x/net/websocket"
)

var onlineTabs map[*websocket.Conn]*Tab

type Tab struct {
	FileName string
	Over     chan bool
}

// InitNotifyServ 初始化通知服务器
func InitNotifyServ() http.Handler {
	onlineTabs = make(map[*websocket.Conn]*Tab)
	return websocket.Handler(onReq)
}

// 处理来自浏览器的请求
func onReq(ws *websocket.Conn) {
	fileName := ws.Request().URL.Query().Get("filename")
	key := ws.Request().URL.Query().Get("key")
	// 避免其他请求
	if key != setting.WSConnKey {
		return
	}
	Tab := &Tab{FileName: fileName, Over: make(chan bool)}
	onlineTabs[ws] = Tab
	defer close(Tab.Over)
	//检查掉线
	go handlerConn(ws, Tab)
	//等待结束
	<-Tab.Over
	//释放资源
	delete(onlineTabs, ws)
	ws.Close()

}

// 处理请求
func handlerConn(ws *websocket.Conn, tab *Tab) {
	//等待掉线
	websocket.JSON.Receive(ws, nil)
	tab.Over <- true
}

// NotifyRefresh 通知浏览器刷新
func NotifyRefresh(fileName string) {
	for ws, tab := range onlineTabs {
		if tab.FileName == fileName {
			websocket.JSON.Send(ws, "refresh")

		}
	}
}

// NotifyMultiRefresh 通知多个文件刷新
func NotifyMultiRefresh(files []string) {
	for _, file := range files {
		NotifyRefresh(file)
	}
}

// GetInjectScript 获取注入脚本
func GetInjectScript() string {
	return injectScript
}

//向页面注入的脚本
var injectScript = `
function connectServ(){
var key="` + setting.WSConnKey + `"
var port="` + setting.Port + `"
var wsservpath="` + setting.WSServPath + `"

var fullpath=window.location.href
var filename=fullpath.substring(fullpath.indexOf("/",7)+1)
if(!(filename.endsWith(".html")||filename.endsWith(".htm"))){
    if(filename.endsWith("/")||filename===""){
        filename+="index.html"
    }else{
        filename+="/index.html"
    }
}

var connURL="ws:localhost:"+port+wsservpath+"?filename="+filename+"&"+"key="+key

var ws=new WebSocket(connURL)

ws.onmessage=function(){
    window.location.reload()
    ws.onclose()
}
}
connectServ()
`
