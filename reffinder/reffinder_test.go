package reffinder

import (
	"fmt"
	"testing"

	"reflect"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCSSFinder(t *testing.T) {

	css := `   @import url("xxx")
	            @import "ccc.css"
	            @import "./d.css"
	                    <style>
	            @import "/pic/asdfdasf.css"
	  @import  sa    "bbb.css"
	  @import url("xxx")
	        </style>
			<style>
	@import "./asdfdasf.css"
	(@import\s*\".+[\"$])`
	cf := new(CSSFinder)
	b := []byte(css)
	cf.FindRef(b)
	fmt.Println(cf.refList)
	Convey("测试CSS引用分析", t, func() {
		So(reflect.DeepEqual(cf.refList, []string{"ccc.css", "./d.css", "/pic/asdfdasf.css", "./asdfdasf.css"}), ShouldBeTrue)
	})
	// fmt.Println(FindFileRef(b, `C:\Users\BYONE\Desktop\t\ss\1.html`, `C:\Users\BYONE\Desktop\t\`))
	// fmt.Println(33)
	// fmt.Println(getRealPath(`C:\Users\BYONE\Desktop\t\ss\1.html`, "/pic/../bbb/pic1.png", `C:\Users\BYONE\Desktop\t\`))
}
func TestHTMLFinder(t *testing.T) {
	html := `<html>
		<head>

			<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
			<meta http-equiv="content-type" content="text/html;charset=utf-8">
			<meta content="always" name="referrer">
	        <meta name="theme-color" content="#2932e1">
	        <link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
	        <link rel="icon" sizes="any" mask href="//www.baidu.com/img/baidu.svg">
	        <link rel="search" type="application/opensearchdescription+xml" href="/content-search.xml" title="百度搜索" />
            <script type="text/javascript" src="myscripts.js"></script>
            <script type="text/javascript" src="http://a.com/myscripts.js"></script>
            <script type="text/javascript" src="//a.com/myscripts.js"></script>
	        <link rel="stylesheet"  href="style.css" type="text/css" />
	        <style>
	            @import url("xxx")
	            @import "ccc.css"
	            @import "./d.css"
	            @import "/pic/asdfdasf.css"
	  @import  sa    "bbb.css"
	  @import url("xxx")
	        </style>
			<style>
	@import "./asdfdasf.css"
	(@import\s*\".+[\"$])
	        </style>
			<img src="sdsadsa.jpg">
	<title>goquery_百度搜索</title>
	</head>
    </html>
	`
	hf := new(HTMLFinder)
	b := []byte(html)
	hf.FindRef(b)
	fmt.Println(hf.refList)
	Convey("测试HTML引用分析", t, func() {
		So(reflect.DeepEqual(hf.refList, []string{"myscripts.js", "http://a.com/myscripts.js", "//a.com/myscripts.js", "style.css", "ccc.css", "./d.css", "/pic/asdfdasf.css", "./asdfdasf.css", "sdsadsa.jpg"}), ShouldBeTrue)
	})
}

func TestFindFileRef(t *testing.T) {
	html := `<html>
		<head>

			<meta http-equiv="X-UA-Compatible" content="IE=edge,chrome=1">
			<meta http-equiv="content-type" content="text/html;charset=utf-8">
			<meta content="always" name="referrer">
	        <meta name="theme-color" content="#2932e1">
	        <link rel="shortcut icon" href="/favicon.ico" type="image/x-icon" />
	        <link rel="icon" sizes="any" mask href="//www.baidu.com/img/baidu.svg">
	        <link rel="search" type="application/opensearchdescription+xml" href="/content-search.xml" title="百度搜索" />
            <script type="text/javascript" src="../myscripts.js"></script>
            <script type="text/javascript" src="http://a.com/myscripts.js"></script>
            <script type="text/javascript" src="//a.com/myscripts.js"></script>
	        <link rel="stylesheet"  href="style.css" type="text/css" />
	        <style>
	            @import url("xxx")
	            @import "ccc.css"
	            @import "./d.css"
	            @import "/pic/asdfdasf.css"
	  @import  sa    "bbb.css"
	  @import url("xxx")
	        </style>
			<style>
	@import "./asdfdasf.css"
	(@import\s*\".+[\"$])
	        </style>
			<img src="sdsadsa.jpg">
	<title>goquery_百度搜索</title>
	</head>
    </html>
	`
	filepath := `~/code/go/src/github.com/thewinds/biu/reffinder/a.html`
	rootpath := `~/code/go/src/github.com/thewinds/biu`
	b := []byte(html)
	refList, err := FindFileRef(b, filepath, rootpath)
	fmt.Println(refList)
	Convey("测试HTML引用分析", t, func() {
		So(err, ShouldBeNil)
		So(reflect.DeepEqual(refList, []string{"myscripts.js", "reffinder/style.css", "reffinder/ccc.css", "reffinder/d.css", "pic/asdfdasf.css", "reffinder/asdfdasf.css", "reffinder/sdsadsa.jpg"}), ShouldBeTrue)
	})
}
