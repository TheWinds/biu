package filerefmap

//FileType 文件类型
type FileType int

const (
	_ FileType = iota
	// JSFile Js文件
	JSFile
	// CSSFile Css文件
	CSSFile
	// HTMLFile Html文件
	HTMLFile
	// IMGFile 图片文件
	IMGFile
	//NotSupportFile 不支持的文件
	NotSupportFile
)
