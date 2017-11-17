package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"regexp"
	"time"
)

type File struct {
	Size int64
	path string
	Name string
}

func (this File) GetSize() int64 {
	return this.Size
}

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
				file.Name(),
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
	nameRegexp regexp.Regexp
}

var onlyTenBiggerFile onlyNBiggerFile

func (this *onlyNBiggerFile) appendFile(file File){
	if !this.nameRegexp.MatchString(file.Name){
		return
	}

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
	begin := time.Now()

	var nameRegexpAsStr string
	flag.StringVar(&nameRegexpAsStr, "nameRegexp", "", "regexp for the file name")

	var path string
	flag.StringVar(&path, "path", "", "Path to folder")

	var minSize int64
	flag.Int64Var(&minSize, "min", 0, "Min size of file")

	var maxSize int64
	flag.Int64Var(&maxSize, "max", -1, "Max size of file (won't be taken into account if negative)")

	//var contentRegexpAsStr string
	//flag.StringVar(&nameRegexpAsStr, "nameRegexp", "", "regexp for the file content")

	flag.Parse() //IMPORTANT!!!! DO NOT USE FLAG VARS BEFORE

	fmt.Println(time.Since(begin))

	nameRegexp, err := regexp.Compile(nameRegexpAsStr)

	if (err != nil){
		panic(err)
	}

	onlyTenBiggerFile = onlyNBiggerFile{10, make([]File, 0), minSize, maxSize, *nameRegexp}

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
				file.Name(),
			})
		} else {
			parseFolderFileChildren(file, path)
		}
	}

	fmt.Println(onlyTenBiggerFile, time.Since(begin))
}
