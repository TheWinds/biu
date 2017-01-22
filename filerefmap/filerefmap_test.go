package filerefmap

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestAll(t *testing.T) {
	// http.ListenAndServe(":8080", http.FileServer(http.Dir("")))
	relmap := &FileRefMap{}
	relmap.AddFile("index", "/index.html", HTMLFile)
	relmap.AddFile("HELLO", "/hello.html", HTMLFile)
	relmap.AddFile("j1", "/js/j1.js", JSFile)
	relmap.AddFile("j2", "/js/j2.js", JSFile)
	relmap.AddFile("j3", "/js/j3.js", JSFile)
	Convey("测试新增文件", t, func() {
		So(relmap.Contains("/index.html"), ShouldBeTrue)
		So(relmap.Contains("/hello.html"), ShouldBeTrue)
		So(relmap.Contains("/js/j1.js"), ShouldBeTrue)
		So(relmap.Contains("/js/j2.js"), ShouldBeTrue)
		So(relmap.Contains("/js/j4.js"), ShouldBeFalse)
	})
	relmap.AddFile("j5", "/js/j5.js", JSFile)
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
		(relmap.AddRef("/js/j3.js", "/js/j2.js"))
		So(relmap.checkCircularRef(), ShouldBeTrue)
	})
	//relmap.ShowInfo()
	//relmap.RemoveFile("/js/j2.js")
	//relmap.ShowInfo()

	//fmt.Println(relmap, relmap.copy())
	// fmt.Println(relmap.FindRoots("/js/j1.js"))
	// fmt.Println(relmap.FindRoots("/js/j3.js"))
	// fmt.Println(relmap.FindRoots("/js/j2.js"))

	fmt.Println(relmap.checkCircularRef())

}
