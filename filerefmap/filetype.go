package filerefmap

type FileType int

const (
	_ FileType = iota
	// JSFile Js文件
	JSFile
	// CSSFile Css文件
	CSSFile
	// HTMLFile HtmlFile
	HTMLFile
)
