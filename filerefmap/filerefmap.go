package filerefmap

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strings"
)

//RefRelNode 文件引用关系节点
type RefRelNode struct {
	FileId  int         //对应文件Id
	NextRef *RefRelNode //下一个节点
}

//FileRefMap 文件关系图
type FileRefMap struct {
	ids    int               //当前id编号
	files  map[int]*FileNode //id->filenode
	pathID map[string]int    //filePath->fileId
}

//根据文件路径取回文件节点
func (frm *FileRefMap) getFileFromPath(filePath string) *FileNode {
	id := frm.pathID[filePath]
	return frm.files[id]

}

//Contains 是否存在文件
func (frm *FileRefMap) Contains(filePath string) bool {
	if frm.pathID == nil {
		return false
	}
	_, contains := frm.pathID[filePath]
	return contains
}

//IsRefFile IsRefFile(A,B) B是否被A引用
func (frm *FileRefMap) IsRefFile(pathFrom, pathTo string) bool {
	if !(frm.Contains(pathFrom) && frm.Contains(pathTo)) {
		return false
	}
	fileFrom := frm.getFileFromPath(pathFrom)
	fileTo := frm.getFileFromPath(pathTo)
	return fileFrom.IsRef(fileTo)
}
func (frm *FileRefMap) copy() *FileRefMap {

	newFiles := make(map[int]*FileNode)
	newPathID := make(map[string]int)
	ids := frm.ids
	deepCopy(&newFiles, frm.files)
	deepCopy(&newPathID, frm.pathID)
	relcpy := &FileRefMap{files: newFiles, pathID: newPathID, ids: ids}
	return relcpy
}

// AddFile 新增文件
func (frm *FileRefMap) AddFile(fileNode FileNode) error {
	//检查path

	//如果当前不存在任何文件,则初始化
	if frm.files == nil {
		frm.files = make(map[int]*FileNode)
		frm.pathID = make(map[string]int)
	}
	if frm.Contains(fileNode.Path) {
		return errors.New("FileRefMap.AddFile:文件已经存在")
	}
	fileType := fileNode.getFileType()
	if fileType == NotSupportFile {
		return errors.New("FileRefMap.AddFile:文件类型不支持")
	}
	//生成文件ID
	fileNode.Id = frm.ids + 1
	fileNode.Type = fileType
	//文件路径=>文件id
	frm.pathID[fileNode.Path] = fileNode.Id
	//id=>文件
	frm.files[fileNode.Id] = &fileNode
	frm.ids++
	return nil
}

// RemoveFile 删除文件
func (frm *FileRefMap) RemoveFile(path string) error {
	//检查path
	//
	if !frm.Contains(path) {
		return errors.New("FileRefMap.RemoveFile:文件不存在")
	}
	//取到文件
	fileNode := frm.getFileFromPath(path)
	if fileNode.RefedCnt != 0 {
		p := fileNode.RefedFiles.NextRef
		for p != nil {
			//删除引用的文件
			frm.files[p.FileId].DelFileRef(frm.files[fileNode.Id])
			p = p.NextRef
		}
	}
	if fileNode.RefCnt != 0 {
		p := fileNode.RefFiles.NextRef
		for p != nil {
			//删除入度
			frm.files[fileNode.Id].DelFileRef(frm.files[p.FileId])
			p = p.NextRef
		}
	}
	delete(frm.files, fileNode.Id)
	delete(frm.pathID, fileNode.Path)
	return nil
}

// RemoveDirFile 根据dir删除文件
func (frm *FileRefMap) RemoveDirFile(dir string) {
	//检查path
	//
	for filePath := range frm.pathID {
		if strings.HasPrefix(filePath, dir) {
			frm.RemoveFile(filePath)
		}
	}
}

// ReNameFile 修改文件名
func (frm *FileRefMap) ReNameFile(oldPath, newPath string) error {
	if !frm.Contains(oldPath) {
		return errors.New("FileRefMap.ReNameFile:文件不存在")
	}
	oldFile := frm.getFileFromPath(oldPath)
	oldFile.Path = newPath
	newType := oldFile.getFileType()
	if newType == NotSupportFile {
		oldFile.Path = oldPath
		return errors.New("FileRefMap.ReNameFile:文件类型不支持")
	}
	oldFile.Path = newPath
	oldFile.Type = oldFile.getFileType()
	frm.pathID[newPath] = oldFile.Id
	delete(frm.pathID, oldPath)
	return nil
}

// AddRef 引用文件
func (frm *FileRefMap) AddRef(pathFrom, pathTo string) error {
	if !(frm.Contains(pathFrom) && frm.Contains(pathTo)) {
		return errors.New("FileRefMap.AddRef:引用文件失败文件不存在")
	}

	fileFrom := frm.getFileFromPath(pathFrom)
	fileTo := frm.getFileFromPath(pathTo)
	if fileFrom.IsRef(fileTo) {
		return errors.New("FileRefMap.AddRef:无须重复引用")
	}
	fileFrom.AddFileRef(fileTo)
	if frm.checkCircularRef() {
		fileFrom.DelFileRef(fileTo)
		return errors.New("FileRefMap.AddRef:存在循环引用")
	}
	return nil
}

// DelRef 删除文件引用
func (frm *FileRefMap) DelRef(pathFrom, pathTo string) error {
	if !(frm.Contains(pathFrom) && frm.Contains(pathTo)) {
		return errors.New("FileRefMap.DelRef:引用文件失败文件不存在")
	}
	fileFrom := frm.getFileFromPath(pathFrom)
	fileTo := frm.getFileFromPath(pathTo)
	fileFrom.DelFileRef(fileTo)
	return nil
}

// UpdateRef 更新新文件引用
func (frm *FileRefMap) UpdateRef(pathFrom string, newRefList []string) []string {
	refs := make([]string, 0, len(newRefList))
	if !frm.Contains(pathFrom) {
		return nil
	}
	fileFrom := frm.getFileFromPath(pathFrom)
	if fileFrom.RefCnt != 0 {
		refedNodes := make([]*FileNode, 0)
		p := fileFrom.RefFiles.NextRef
		//寻找多余引用
		for p != nil {
			refedNodes = append(refedNodes, frm.files[p.FileId])
			p = p.NextRef
		}
		fileFrom.ReSetFileRef(refedNodes)
	}
	for _, newRef := range newRefList {
		if frm.Contains(newRef) {
			fileTo := frm.getFileFromPath(newRef)
			if !fileFrom.IsRef(fileTo) {
				fileFrom.AddFileRef(fileTo)
				refs = append(refs, newRef)
			}
		}
	}
	return refs
}

//FindRoots 寻找根节点
func (frm *FileRefMap) FindRoots(filePath string) []string {
	if !frm.Contains(filePath) {
		return nil
	}
	fileTo := frm.getFileFromPath(filePath)
	var roots []string
	if fileTo.Type == HTMLFile {
		return []string{fileTo.Path}
	}
	for _, fileNode := range frm.files {
		if fileNode.Type != HTMLFile {
			continue
		}
		//
		if frm.isConnect(fileNode, fileTo) {
			roots = append(roots, fileNode.Path)
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
				//把当前节点加入队列
				queue <- frm.files[p.FileId]
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
