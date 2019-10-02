package main

import (
	"os"
	"path/filepath"
	"strings"
	"fmt"
	"io/ioutil"
	"io"
)

func getTreeFileList(filePath string, Files bool) ([]string, error) {
	var fileList []string

	err := filepath.Walk(filePath, func(path string, f os.FileInfo, err error) error {
		if isDisabled(path) {
			return nil
		}

		if !f.IsDir() && !Files {
			return nil
		}

		fileList = append(fileList, path)
		return nil
	})

	return fileList, err
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("go run main.go -f")
	}
	path := "."
	printFiles := (len(os.Args) == 2 && os.Args[1] == "-f")
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
func dirTree(out io.Writer, filePath string, printFiles bool) error {
	var resultTree string
	var GlobalFiles = printFiles

	fileList, err := getTreeFileList(filePath, GlobalFiles)

	for index, file := range fileList {
		treeLine := getLinePath(file, GlobalFiles)
		if treeLine == "" {
			continue
		}

		resultTree = resultTree + treeLine

		if (len(fileList) - 1) != index {
			resultTree = resultTree + "\n"
		}
	}

	fmt.Fprintln(out, resultTree)

	return err
}

func getLinePath(path string, gfiles bool) string {
	var pathResult string
	var tabs string
	pathLinux := strings.Replace(path, `\`, `/`, 100)
	pathListFull := strings.Split(pathLinux, `/`)

	pathList := pathListFull[1:]

	if len(pathList) == 0 {
		return pathResult
	}

	basePath := filepath.Base(path) + FileSize(path)

	if isLastElementPath(path, gfiles) {
		pathResult = pathResult + `└───` + basePath
	} else {
		pathResult = pathResult + `├───` + basePath
	}

	tabs = getTabs(pathListFull, gfiles)

	return tabs + pathResult
}

func getTabs(pathList []string, gfiles bool) string {
	var tabResult string

	for i := 2; i < len(pathList); i++ {
		if isLastElementPath(filepath.Join(pathList[:i]...), gfiles) {
			tabResult = tabResult + "\t"
		} else {
			tabResult = tabResult + "│\t"
		}
	}

	return tabResult
}

func isLastElementPath(path string, Gfiles bool) bool {

	basePath := filepath.Base(path)

	var sortList []string

	files, _ := ioutil.ReadDir(filepath.Dir(path))

	for _, file := range files {
		if Gfiles == false && file.IsDir() == false {
			continue
		}
		sortList = append(sortList, file.Name())
	}

	if sortList[len(sortList)-1] == basePath {
		return true
	}

	return false
}

func FileSize(path string) string {
	var fileSize string
	fileInfo, _ := os.Stat(path)
	if !fileInfo.IsDir() {
		size := fileInfo.Size()
		if size == 0 {
			fileSize = " (empty)"
		} else {
			fileSize = fmt.Sprintf(" (%vb)", size)
		}
	}

	return fileSize
}

func isDisabled(path string) bool {
	disabledFound := []string{".git", ".gitignore", ".idea", "README.md", ".", "test_compare"}
	pathList := strings.Split(path, `\`)

	for _, value := range disabledFound {
		if pathList[0] == value {
			return true
		}
	}
	return false
}
