package filerefmap

import "errors"

//FileNode 文件节点
type FileNode struct {
	Name string   //文件名
	Path string   //文件路径
	Type FileType //文件类型

	RefCnt   int         //引用文件计数
	RefFiles *RefRelNode //引用文件列表

	RefedCnt   int         //被引用文件计数
	RefedFiles *RefRelNode //被引的文件列表
}

//Show 显示节点信息
// func (node *FileNode) show() {
// 	fmt.Println("Name:", node.Name)
// 	fmt.Println("Path:", node.Path)
// 	fmt.Println("Type:", node.Type)
// 	fmt.Printf("RefedFiles:\n")
// 	if node.RefedCnt != 0 {
// 		p := node.RefedFiles.NextRef
// 		for p != nil {
// 			fmt.Print(p.NodePath, " ")
// 			p = p.NextRef
// 		}
// 	}
// 	if node.RefCnt != 0 {
// 		fmt.Printf("\nRefFiles:\n")
// 		pp := node.RefFiles.NextRef
// 		for pp != nil {
// 			fmt.Print(pp.NodePath, " ")
// 			pp = pp.NextRef
// 		}
// 	}
// 	fmt.Println()

// }

//AddFileRef 新增文件引用 参数（被引用的节点）
func (node *FileNode) AddFileRef(refedNode *FileNode) {
	//将被引用的节点加入当前节点的引用列表
	node.addRefAndChangeCnt(refedNode, &node.RefFiles, &node.RefCnt)
	//将当前节点加入被引用的节点的被引用列表
	node.addRefAndChangeCnt(node, &refedNode.RefedFiles, &refedNode.RefedCnt)
}

//DelFileRef 删除文件引用 参数（被删除的文件路径）
func (node *FileNode) DelFileRef(refedNode *FileNode) {
	//将被引用的节点从当前引用列表中删除
	node.delRefAndChangeCnt(refedNode.Path, &node.RefFiles, &node.RefCnt)
	//将被当前节点从被引用节点的被引用列表中删除
	node.delRefAndChangeCnt(node.Path, &refedNode.RefedFiles, &refedNode.RefedCnt)
}

//addRefAndChangeCnt 新增引用并且计数
func (node *FileNode) addRefAndChangeCnt(refedNode *FileNode, refTo **RefRelNode, cnt *int) {
	//在当前节点将被引用的节点加入引用列表
	newRef := &RefRelNode{NodePath: refedNode.Path}
	if *cnt == 0 {
		*refTo = new(RefRelNode)
	}
	newRef.NextRef = (*refTo).NextRef
	(*refTo).NextRef = newRef
	*cnt++
}

//delRefAndChangeCnt 删除引用并且计数
func (node *FileNode) delRefAndChangeCnt(refedFilePath string, refTo **RefRelNode, cnt *int) error {

	var curRef = (*refTo).NextRef
	var frontCurRef = (*refTo)
	isFind := false
	for curRef != nil {
		if curRef.NodePath == refedFilePath {
			frontCurRef.NextRef = curRef.NextRef
			isFind = true
			break
		}
		curRef = curRef.NextRef
		frontCurRef = frontCurRef.NextRef
	}
	if !isFind {
		return errors.New("FileNode:边不存在")
	}
	*cnt--
	return nil
}
