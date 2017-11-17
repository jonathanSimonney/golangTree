package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
)

//type IFileOrFolder interface{
//	GetSize() int64
//}

type File struct {
	Size int64
	path string
}
//type Folder struct {
//	Elems []IFileOrFolder
//}

func (this File) GetSize() int64 {
	return this.Size
}

//func (this Folder) GetSize() int64 {
//	var totalSize int64
//	totalSize = 0
//	for _, fileOrFolder := range this.Elems {
//		totalSize += fileOrFolder.GetSize()
//	}
//
//	return totalSize
//}

//func getFolderChildren(inputFile os.FileInfo, fileLoc string) []IFileOrFolder {
//	filePath := fileLoc + "/" + inputFile.Name()
//
//	var children []IFileOrFolder
//	files, err := ioutil.ReadDir(filePath)
//	if err != nil {
//		panic(err)
//	}
//
//	for _, file := range files {
//		if !file.IsDir(){
//			children = append(children, File{
//				file.Size(),
//			})
//		}else{
//			children = append(children, Folder{
//				getFolderChildren(file, filePath),
//			})
//		}
//	}
//
//	return children
//}

func parseFolderFileChildren(inputFolder os.FileInfo, fileLoc string){
	filePath := fileLoc + "/" + inputFolder.Name()

	files, err := ioutil.ReadDir(filePath)
	if err != nil {
		if err.Error() == "open C:\\/Documents and Settings: Accès refusé." || err.Error() == "open C:\\/Program Files/Fichiers communs: Accès refusé."{
			fmt.Println("normal")
		}else{
			fmt.Println(err)
			fmt.Println("hope it wasn't too important...")
		}
	}

	for _, file := range files {
		if !file.IsDir(){
			onlyTenBiggerFile.appendFile(File{
				file.Size(),
				filePath + "/" + file.Name(),
			})
		}else{
			parseFolderFileChildren(file, filePath)
		}
	}
}

type onlyNBiggerFile struct{
	n int
	files []File
	min int64
	max int64
}

var onlyTenBiggerFile onlyNBiggerFile

func (this *onlyNBiggerFile) appendFile(file File){
	if file.Size <= this.min || (this.max > 0 && file.Size >= this.max){
		return
	}

	if len(this.files) < this.n{
		this.files = append(this.files, file)
	}else{
		currentMinSize := this.files[this.n - 1].Size
		candidateSize := file.Size

		if candidateSize > currentMinSize{
			this.files[this.n - 1] = file
		}
	}

	sort.SliceStable(this.files, func(i, j int) bool{return this.files[i].Size > this.files[j].Size})
}

func main() {
	var path string
	flag.StringVar(&path, "path", "", "Path to folder")

	var minSize int64
	flag.Int64Var(&minSize, "min", 0, "Min size of file")

	var maxSize int64
	flag.Int64Var(&maxSize, "max", -1, "Max size of file (won't be taken into account if negative)")

	flag.Parse()
	onlyTenBiggerFile = onlyNBiggerFile{10, make([]File, 0), minSize, maxSize}

	files, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Println("Please enter a valid path to a directory")
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			onlyTenBiggerFile.appendFile(File{
				file.Size(),
				path + "/" + file.Name(),
			})
		} else {
			parseFolderFileChildren(file, path)
		}
	}

	fmt.Println(onlyTenBiggerFile)
}
