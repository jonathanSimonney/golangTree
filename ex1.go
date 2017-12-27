package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"regexp"
	"time"
	"runtime"
	"sync"
	"bufio"
	"log"
	"runtime/pprof"
)

type File struct {
	Size int64
	path string
	Name string
}

func (this File) GetSize() int64 {
	return this.Size
}

func parseFolderFileChildren(inputFolder os.FileInfo, fileLoc string, wg *sync.WaitGroup, beginningChan chan File){
	filePath := fileLoc + "/" + inputFolder.Name()

	files, _ := ioutil.ReadDir(filePath)
	//if err != nil {
	//	if err.Error() == "open C:\\/Documents and Settings: Accès refusé." || err.Error() == "open C:\\/Program Files/Fichiers communs: Accès refusé."{
	//		fmt.Println("normal")
	//	}else{
	//		fmt.Println(err)
	//		fmt.Println("hope it wasn't too important...")
	//	}
	//}

	for _, file := range files {
		if !file.IsDir(){
			onlyTenBiggerFile.appendFile(File{
				file.Size(),
				filePath + "/" + file.Name(),
				file.Name(),
			}, wg, beginningChan)
		}else{
			parseFolderFileChildren(file, filePath, wg, beginningChan)
		}
	}
}

type onlyNBiggerFile struct{
	n int
	files []File
	min int64
	max int64
	nameRegexp regexp.Regexp
	contentRegexp regexp.Regexp
}

var onlyTenBiggerFile onlyNBiggerFile

func (this *onlyNBiggerFile) appendFile(file File, wg *sync.WaitGroup, beginningChan chan File){
	wg.Add(1)

	beginningChan <- file
}

func addFile(c <- chan File, wg *sync.WaitGroup, nextChan chan File){
	for{
		file := <- c // blocking
		if !onlyTenBiggerFile.nameRegexp.MatchString(file.Name){
			//fmt.Println("failed to match regexp : file name : " + file.Name)
			wg.Done()
			continue
		}

		//fmt.Println("successed to match regexp : file name : " + file.Name)

		nextChan <- file
	}
}

func checkFileSize(c <- chan File, wg *sync.WaitGroup, nextChan chan File){
	for{
		file := <- c // blocking

		if file.Size <= onlyTenBiggerFile.min || (onlyTenBiggerFile.max > 0 && file.Size >= onlyTenBiggerFile.max){
			wg.Done()
			continue
		}

		nextChan <- file
	}
}

func checkFileContent(c <- chan File, wg *sync.WaitGroup, nextChan chan File){
	for{
		file := <- c // blocking

		readFile, err := os.Open(file.path)
		if err != nil{
			fmt.Println("file " + file.path + " couldn't be read!")
			wg.Done()
			continue
		}

		reader := bufio.NewReader(readFile)
		firstHundredChar, err := reader.Peek(100)
		if err != nil{
			//fmt.Println(err)
		}

		if !onlyTenBiggerFile.contentRegexp.MatchString(string(firstHundredChar)){
			wg.Done()
			continue
		}

		nextChan <- file
	}
}

func addFileToList(c <- chan File, wg *sync.WaitGroup, onlyTenBiggerFile *onlyNBiggerFile){
	for{
		file := <- c // blocking

		if len(onlyTenBiggerFile.files) < onlyTenBiggerFile.n{
			onlyTenBiggerFile.files = append(onlyTenBiggerFile.files, file)
		}else{
			currentMinSize := onlyTenBiggerFile.files[onlyTenBiggerFile.n - 1].Size
			candidateSize := file.Size

			if candidateSize > currentMinSize{
				onlyTenBiggerFile.files[onlyTenBiggerFile.n - 1] = file
			}
		}

		sort.SliceStable(onlyTenBiggerFile.files, func(i, j int) bool{return onlyTenBiggerFile.files[i].Size > onlyTenBiggerFile.files[j].Size})

		wg.Done()
	}
}

func programMain() {
	begin := time.Now()
	nbCores := runtime.NumCPU()
	wg := sync.WaitGroup{}

	var nameRegexpAsStr string
	flag.StringVar(&nameRegexpAsStr, "nameRegexp", "", "regexp for the file name")

	var path string
	flag.StringVar(&path, "path", "", "Path to folder")

	var minSize int64
	flag.Int64Var(&minSize, "min", 0, "Min size of file")

	var maxSize int64
	flag.Int64Var(&maxSize, "max", -1, "Max size of file (won't be taken into account if negative)")

	var contentRegexpAsStr string
	flag.StringVar(&contentRegexpAsStr, "contentRegexp", "", "regexp for the file content")

	flag.Parse() //IMPORTANT!!!! DO NOT USE FLAG VARS BEFORE

	fmt.Println(time.Since(begin))

	nameRegexp, err := regexp.Compile(nameRegexpAsStr)

	if err != nil{
		panic(err)
	}

	contentRegexp, err := regexp.Compile(contentRegexpAsStr)

	if err != nil{
		panic(err)
	}

	onlyTenBiggerFile = onlyNBiggerFile{10, make([]File, 0), minSize, maxSize, *nameRegexp, *contentRegexp}

	beginningChan := make(chan File)
	chanInter1 := make(chan File)
	chanInter2 := make(chan File)
	chanFinal := make(chan File)
	for i := 0; i < nbCores; i++{
		go addFile(beginningChan, &wg, chanInter1)
		go checkFileSize(chanInter1, &wg, chanInter2)
		go checkFileContent(chanInter2, &wg, chanFinal)
	}

	go addFileToList(chanFinal, &wg, &onlyTenBiggerFile)

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
			}, &wg, beginningChan)
		} else {
			parseFolderFileChildren(file, path, &wg, beginningChan)
		}
	}

	wg.Wait()

	fmt.Println(onlyTenBiggerFile, time.Since(begin))
}

func main(){
	f, err := os.Create("perf_cpu.perf")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	defer pprof.StopCPUProfile()

	programMain()


	f, err = os.Create("mem_profile.perf")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	runtime.GC() // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}
	f.Close()


}
