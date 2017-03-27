package main

import (
	"bytes"
	"net/http"
	"strings"

	"fmt"

	"github.com/thewinds/biu/setting"
	"golang.org/x/net/websocket"
)

var onlineTabs map[*websocket.Conn]*Tab

// Tab 页面标签
type Tab struct {
	FileName string
	Over     chan bool
}

// InitNotifyServ 初始化通知刷新服务器
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
			RefreshTimes++
			updateTerm()
		}
	}
}

// NotifyMultiRefresh 通知多个文件刷新
func NotifyMultiRefresh(files []string) {
	for _, file := range files {
		NotifyRefresh(file)
	}
}

// InjectScriptHandler 注入脚本的handler
func InjectScriptHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Write(buildInjectScript(setting.WSServPath, setting.Port, setting.WSConnKey))
}

// InjectScriptFunc 向页面注入内容的func
func InjectScriptFunc(name string, basebytes []byte) []byte {
	if strings.HasSuffix(name, ".html") {
		return bytes.Replace(
			basebytes,
			[]byte(`</body>`),
			[]byte(`<script src="`+setting.InjectScriptPath+`"></script>`+"\n"+"</body>"),
			1)
	}
	return basebytes
}

//构建向页面注入的脚本
func buildInjectScript(servPath, port, key string) []byte {
	return []byte(fmt.Sprintf(injectScriptTemplate, key, port, servPath))
}

//注入脚本模板
const injectScriptTemplate = `
function connectServ(){
var key="%s"
var port=%s
var wsservpath="%s"

var pathname=window.location.pathname

if (pathname==="/"){
    pathname+="index.html"
}
pathname=pathname.substring(1)

var connURL="ws:localhost:"+port+wsservpath+"?filename="+pathname+"&"+"key="+key

var ws=new WebSocket(connURL)

ws.onmessage=function(){
    window.location.reload()
    ws.close()
}
}
connectServ()
`
