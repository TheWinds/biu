package reffinder

import (
	"errors"
	"strings"

	"regexp"

	"bytes"

	"runtime"

	"path/filepath"

	"github.com/PuerkitoBio/goquery"
)

var finders = map[string]interface{}{
	".js":   new(JSFinder),
	".css":  new(CSSFinder),
	".html": new(HTMLFinder),
}

//FindFileRef 查找文件引用
func FindFileRef(file []byte, filePath, rootPath string) ([]string, error) {
	//获取文件类型
	ext := filepath.Ext(filePath)
	ext = strings.ToLower(ext)
	//查找是否存在对应文件类型的引用发现器
	finder, has := finders[ext]
	if !has {
		return nil, errors.New("FindFileRef:文件类型不支持,(type):" + ext)
	}
	//查找引用的文件转换为真实路径
	(finder.(refFinder)).Reset()
	refList := (finder.(refFinder)).FindRef(file)
	var ret []string
	for _, refFilePath := range refList {
		//过滤掉非法路径
		if refFilePath != "" && !isURL(refFilePath) {
			realPath, err := getRealPath(filePath, refFilePath, rootPath)
			if err != nil {
				return nil, err
			}
			ret = append(ret, realPath)
		}
	}
	return ret, nil
}

//getRealPath 获取被引用文件相对根目录的真实地址
func getRealPath(filePath, refFilePath, rootPath string) (string, error) {
	pathSep := "/"
	if runtime.GOOS == "windows" {
		pathSep = "\\"
	}
	//统一格式
	if !strings.HasSuffix(rootPath, pathSep) {
		rootPath = rootPath + pathSep
	}
	if strings.HasPrefix(refFilePath, "/") {
		refFilePath = "~" + refFilePath
	}
	//取到文件相对与网站根目录的路径
	relRootPath := strings.Replace(filePath, rootPath, "", -1)
	//分割剩下目录信息
	relRootPaths := strings.Split(relRootPath, pathSep)
	// fmt.Println(relRootPath, "：", relRootPaths, len(relRootPaths))
	relRootPaths = relRootPaths[:len(relRootPaths)-1]
	//分割引用文件路径剩下目录和文件信息
	refFilePaths := strings.Split(refFilePath, "/")
	var newPaths []string
	if refFilePaths[0] == "." {
		newPaths = relRootPaths
		refFilePaths = refFilePaths[1:]
	} else if refFilePaths[0] == "~" {
		refFilePaths = refFilePaths[1:]
	} else {
		newPaths = relRootPaths
	}
	//查找真实路径
	for _, token := range refFilePaths {
		if token == ".." {
			if len(newPaths) == 0 {
				return "", errors.New("引用路径错误")
			}
			newPaths = newPaths[:len(newPaths)-1]
		} else {
			newPaths = append(newPaths, token)
		}
	}
	realPath := ""
	//还原路径
	for i, path := range newPaths {
		if i != 0 {
			realPath += pathSep
		}
		realPath += path
	}
	return realPath, nil
}

//refFinder 医用发现器接口
type refFinder interface {
	FindRef(inputs []byte) []string
	Reset()
}

// HTMLFinder 引用发现器
type HTMLFinder struct {
	refList []string
}

// Reset 重置
func (hf *HTMLFinder) Reset() {
	hf.refList = make([]string, 20)
}

//FindRef 查找HTML文件的所有引用
func (hf *HTMLFinder) FindRef(inputs []byte) []string {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(inputs))
	if err != nil {
		return nil
	}
	doc.Find("script").Each(hf.queryJSFromScript)
	doc.Find("link").Each(hf.queryCSSFromLink)
	doc.Find("style").Each(hf.queryCSSFromStyle)
	doc.Find("img").Each(hf.queryImgFromImg)
	return hf.refList
}
func (hf *HTMLFinder) queryCSSFromLink(i int, s *goquery.Selection) {
	if rel, exist := s.Attr("rel"); !(exist && rel == "stylesheet") {
		return
	}
	if href, exist := s.Attr("href"); exist {
		hf.refList = append(hf.refList, href)
	}
}
func (hf *HTMLFinder) queryCSSFromStyle(i int, s *goquery.Selection) {
	src := s.Text()
	if !strings.Contains(src, "@import") {
		return
	}
	//正则匹配在css中引用css的代码
	patternImportCSSLine := `@import\s*\".+\.css\"`
	matcher, _ := regexp.Compile(patternImportCSSLine)
	//匹配所有
	importcss := matcher.FindAllString(src, -1)
	//提取引用的文件
	for index := 0; index < len(importcss); index++ {
		if strings.HasSuffix(importcss[index], "\"") {
			start := strings.Index(importcss[index], "\"")
			end := strings.LastIndex(importcss[index], "\"")
			if start < end {
				importcss[index] = importcss[index][start+1 : end]
			} else {
				importcss[index] = ""
			}

		}
	}
	//返回结果
	if len(importcss) != 0 {
		hf.refList = append(hf.refList, importcss...)
	}
}
func (hf *HTMLFinder) queryImgFromImg(i int, s *goquery.Selection) {
	if src, exist := s.Attr("src"); exist {
		hf.refList = append(hf.refList, src)
	}
}
func (hf *HTMLFinder) queryJSFromScript(i int, s *goquery.Selection) {
	if src, exist := s.Attr("src"); exist {
		hf.refList = append(hf.refList, src)
	}
}

// JSFinder 引用发现器
type JSFinder struct {
	refList []string
}

// Reset 重置
func (jf *JSFinder) Reset() {
	jf.refList = make([]string, 20)
}

//FindRef 查找HTML文件的所有引用
func (jf *JSFinder) FindRef(inputs []byte) []string {
	//尚未实现
	return nil
}

// CSSFinder 引用发现器
type CSSFinder struct {
	refList []string
}

// Reset 重置
func (cf *CSSFinder) Reset() {
	cf.refList = make([]string, 20)
}

//FindRef 查找HTML文件的所有引用
func (cf *CSSFinder) FindRef(inputs []byte) []string {
	src := string(inputs)
	if !strings.Contains(src, "@import") {
		return nil
	}
	//正则匹配在css中引用css的代码
	patternImportCSSLine := `@import\s*\".+\.css\"`
	matcher, _ := regexp.Compile(patternImportCSSLine)
	//匹配所有
	importcss := matcher.FindAllString(src, -1)
	//提取引用的文件
	for index := 0; index < len(importcss); index++ {
		if strings.HasSuffix(importcss[index], "\"") {
			start := strings.Index(importcss[index], "\"")
			end := strings.LastIndex(importcss[index], "\"")
			if start < end {
				importcss[index] = importcss[index][start+1 : end]
			} else {
				importcss[index] = ""
			}

		}
	}
	//返回结果
	if len(importcss) != 0 {
		cf.refList = append(cf.refList, importcss...)
	}
	return cf.refList
}

//
func isURL(path string) bool {
	pattern := `(http|ftp|https):\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?|\/\/[\w\-_]+(\.[\w\-_]+)+([\w\-\.,@?^=%&:/~\+#]*[\w\-\@?^=%&/~\+#])?`
	matcher, _ := regexp.Compile(pattern)
	return matcher.MatchString(path)

}
