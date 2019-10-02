package main

// сюда писать функцию DirTree

//func dirTree(out io.Writer, filePath string, printFiles bool) error {
//	var resultTree string
//	printFilesGlobal = printFiles
//
//	fileList, err := getTreeFileList(filePath)
//
//	for index, file := range fileList {
//		treeLine := getLinePath(file)
//		if treeLine == "" {
//			continue
//		}
//
//		resultTree = resultTree + treeLine
//
//		if (len(fileList) - 1) != index {
//			resultTree = resultTree + "\n"
//		}
//	}
//
//	fmt.Fprintln(out, resultTree)
//
//	return err
//}