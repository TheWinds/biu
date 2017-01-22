package filerefmap

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAll(t *testing.T) {
	// http.ListenAndServe(":8080", http.FileServer(http.Dir("")))
	relmap := &FileRefMap{}
	files := []FileNode{
		FileNode{Path: "/index.html"},
		FileNode{Path: "/hello.html"},
		FileNode{Path: "/js/j1.js"},
		FileNode{Path: "/js/j2.js"},
		FileNode{Path: "/js/j3.js"},
	}
	for _, file := range files {
		fmt.Println(relmap.AddFile(file))
	}
	Convey("测试新增文件", t, func() {
		So(relmap.Contains("/index.html"), ShouldBeTrue)
		So(relmap.Contains("/hello.html"), ShouldBeTrue)
		So(relmap.Contains("/js/j1.js"), ShouldBeTrue)
		So(relmap.Contains("/js/j2.js"), ShouldBeTrue)
		So(relmap.Contains("/js/j4.js"), ShouldBeFalse)
	})
	relmap.AddFile(FileNode{Path: "/js/j5.js"})
	relmap.RemoveFile("/js/j5.js")
	Convey("测试删除文件", t, func() {
		So(relmap.Contains("/js/j5.js"), ShouldBeFalse)
	})
	//relmap.ShowInfo()
	(relmap.AddRef("/index.html", "/js/j1.js"))
	(relmap.AddRef("/index.html", "/js/j2.js"))
	(relmap.AddRef("/js/j2.js", "/js/j3.js"))
	(relmap.AddRef("/hello.html", "/js/j3.js"))
	Convey("测试寻找根节点", t, func() {
		So(reflect.DeepEqual(relmap.FindRoots("/js/j3.js"), []string{"/index.html", "/hello.html"}), ShouldBeTrue)
	})
	Convey("测试检查循环引用", t, func() {
		So(relmap.checkCircularRef(), ShouldBeFalse)
		So(reflect.DeepEqual(relmap.AddRef("/js/j3.js", "/js/j2.js"), errors.New("FileRefMap.AddRef:存在循环引用")), ShouldBeTrue)
		So(relmap.IsRefFile("/js/j3.js", "/js/j2.js"), ShouldBeFalse)

		// So(relmap.checkCircularRef(), ShouldBeTrue)
	})
	//relmap.ShowInfo()
	//relmap.RemoveFile("/js/j2.js")
	//relmap.ShowInfo()

	//fmt.Println(relmap, relmap.copy())
	// fmt.Println(relmap.FindRoots("/js/j1.js"))
	// fmt.Println(relmap.FindRoots("/js/j3.js"))
	// fmt.Println(relmap.FindRoots("/js/j2.js"))

	fmt.Println(relmap.checkCircularRef())
	relmap.ReNameFile("/index.html", "/index_new.html")
	Convey("测试文件改名", t, func() {
		So(relmap.Contains("/index.html"), ShouldBeFalse)
		So(relmap.Contains("/index_new.html"), ShouldBeTrue)
		So(relmap.IsRefFile("/index_new.html", "/js/j1.js"), ShouldBeTrue)
	})
}
