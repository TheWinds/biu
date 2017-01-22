package filerefmap

import (
	"bytes"
	"encoding/gob"
	"errors"
)

//RefRelNode 文件引用关系节点
type RefRelNode struct {
	NodePath string      //对应文件路径
	NextRef  *RefRelNode //下一个节点
}

//FileRefMap 文件关系图
type FileRefMap struct {
	files map[string]*FileNode
}

//Contains 是否存在文件
func (frm *FileRefMap) Contains(path string) bool {
	_, contains := frm.files[path]
	return contains
}

func (frm *FileRefMap) copy() *FileRefMap {

	newFiles := make(map[string]*FileNode)
	deepCopy(&newFiles, frm.files)
	relcpy := &FileRefMap{files: newFiles}
	return relcpy
}

// AddFile 新增文件
func (frm *FileRefMap) AddFile(name, path string, filetype FileType) error {
	//检查path
	//
	if frm.files == nil {
		frm.files = make(map[string]*FileNode)
	}
	if frm.Contains(path) {
		return errors.New("FileRefMap:文件已经存在")
	}
	node := &FileNode{Name: name, Path: path, Type: filetype}
	frm.files[path] = node
	return nil
}

// RemoveFile 删除文件
func (frm *FileRefMap) RemoveFile(path string) error {
	//检查path
	//
	if frm.files == nil {
		return errors.New("FileRefMap:文件不存在")
	}
	if !frm.Contains(path) {
		return errors.New("FileRefMap:文件不存在")
	}
	node := frm.files[path]
	if node.RefedCnt != 0 {
		p := node.RefedFiles.NextRef
		for p != nil {
			//删除引用的文件
			frm.files[p.NodePath].DelFileRef(frm.files[node.Path])
			p = p.NextRef
		}
	}
	if node.RefCnt != 0 {
		p := node.RefFiles.NextRef
		for p != nil {
			//删除入度
			frm.files[node.Path].DelFileRef(frm.files[p.NodePath])
			p = p.NextRef
		}
	}
	delete(frm.files, path)
	return nil
}

// AddRef 引用文件
func (frm *FileRefMap) AddRef(pathFrom, pathTo string) error {
	if frm.files == nil {
		return errors.New("FileRefMap:引用文件不存在")
	}
	if !(frm.Contains(pathFrom) && frm.Contains(pathTo)) {
		return errors.New("FileRefMap:引用文件失败文件不存在")
	}

	frm.files[pathFrom].AddFileRef(frm.files[pathTo])
	return nil
}

//FindRoots 寻找根节点
func (frm *FileRefMap) FindRoots(filePath string) []string {
	if !frm.Contains(filePath) {
		return nil
	}
	var roots []string
	for _, node := range frm.files {
		if node.Type != HTMLFile {
			continue
		}
		if frm.isConnect(node, frm.files[filePath]) {
			roots = append(roots, node.Path)
		}
	}
	return roots
}

//isConnect 检测两个节点是否连同
func (frm *FileRefMap) isConnect(from *FileNode, to *FileNode) bool {
	queue := make(chan *FileNode, len(frm.files))
	queue <- from
	for len(queue) != 0 {
		frontNode := <-queue
		if frontNode.Path == to.Path {
			return true
		}
		if frontNode.RefCnt != 0 {
			p := frontNode.RefFiles.NextRef
			for p != nil {
				queue <- frm.files[p.NodePath]
				p = p.NextRef
			}
		}
	}

	return false
}

//checkCircularRef 检测是否存在循环引用
func (frm *FileRefMap) checkCircularRef() bool {
	cp := frm.copy()
	for len(cp.files) != 0 {
		findPath := ""
		for _, fileNode := range cp.files {
			if fileNode.RefedCnt == 0 {
				findPath = fileNode.Path
				break
			}
		}
		if findPath == "" {
			return true
		}
		cp.RemoveFile(findPath)
	}
	return false
}

// showInfo 显示引用信息
// func (frm FileRefMap) showInfo() {
// 	for _, node := range frm.files {
// 		fmt.Println("++++++++++++++++++++++++++++++++++++")
// 		node.show()
// 	}
// 	fmt.Println("=============================================")

// }

//deepCopy 拷贝
func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)

}
