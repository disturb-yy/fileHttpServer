/*
************************************************************************
> File Name:     main.go
> Author:        程序员Carl
> 微信公众号:    代码随想录
> Created Time:  Wed May 16 14:30:07 2018
> Description:

	***********************************************************************
*/
package main

import (
	"errors"
	"fileHttpServer/settings"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
)

const COVERVALUE string = "noCover"
const COVERKEY string = "cover"
const PARAMNUMCHECK int = 3

// check the file name exist or not
func checkFileName(fileName string) (bool, string) {
	bExist, err := exists(generateUpload(fileName))
	if err != nil {
		log.Fatal(err)
	}
	// 如果文件已存在，在文件名filename后面加上序号
	if bExist == true {
		version := 1
		bfExist := true
		var newFileName string
		for bfExist {
			var idx int
			for i, ch := range fileName {
				if ch == '.' {
					idx = i
					break
				}
			}
			newFileName = fmt.Sprintf("%s_%d.%s", fileName[:idx], version, fileName[idx+1:])
			bfExist, _ = exists(generateUpload(newFileName))
			version++
		}
		fileName = newFileName
	}

	return bExist, fileName
}
func postMethod(w http.ResponseWriter, r *http.Request) {
	//The whole request body is parsed and up to a total of maxMemory bytes of its file parts are stored in memory
	// 设置内存大小
	err := r.ParseMultipartForm(settings.Conf.MaxMemory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.ParseForm()
	sCover := r.Form.Get(COVERKEY)
	fmt.Printf(COVERKEY+":%s\n", sCover)
	//get a ref to the parsed multipart form
	m := r.MultipartForm
	files := m.File["uploadfile"]
	for i := range files {
		bNewFileName := false
		newFileName := files[i].Filename
		if sCover == COVERVALUE { // 不覆盖模式保存文件，生成新的文件名
			bNewFileName, newFileName = checkFileName(files[i].Filename)
		}
		fmt.Printf("fileNname[%d]:"+files[i].Filename+", newFileName:"+newFileName+"\n", i)

		file, err := files[i].Open()
		defer file.Close()
		// 创建文件
		targetFile, err := os.Create(generateUpload(newFileName))
		defer targetFile.Close()
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(targetFile, file)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		downloadPath := generateDown(newFileName)
		if bNewFileName == true {
			w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\n%s already exists, "+
				"new file name:%s\nGet object way: %s\n",
				n, files[i].Filename, newFileName, downloadPath)))
		} else {
			w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\nGet object way: %s\n",
				n, downloadPath)))
		}
	}
}
func putMethod(w http.ResponseWriter, r *http.Request) {
	fileName, err := getFileNameFromURL(r.URL.Path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("getHandle URL:%s, filename:%s\n", r.URL.Path, fileName)

	// 在 upload 文件夹生成 filename 文件
	targetFile := new(os.File)
	targetFile, err = os.Create(generateUpload(fileName))
	defer targetFile.Close()
	if err != nil {
		panic(err)
	}
	file := r.Body
	n, err := io.Copy(targetFile, file)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("%d bytes are recieved.\nGet object way: %s\n",
		n, generateDown(fileName))))
}

// upload object
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("method:" + r.Method + "\n")
	if r.Method == "PUT" {
		putMethod(w, r)
	}
	if r.Method == "POST" {
		postMethod(w, r)
	}
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	// 文件不存在，返回 false
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return true, err
}

// generate upload link
func generateUpload(filename string) string {
	return fmt.Sprintf("%s/%s", settings.Conf.UploadPath, filename)
}

// generate download link
func generateDown(filename string) string {
	return fmt.Sprintf("wegt http://%s:%d/%s",
		settings.Conf.Host,
		settings.Conf.Port,
		filename)
}

// get object
func getHandle(w http.ResponseWriter, r *http.Request) {
	fileName, err := getFileNameFromURL(r.URL.Path)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("getHandle URL:%s, filename:%s\n", r.URL.Path, fileName)
	objectPath := generateUpload(fileName)
	// 获取 objectPath 的资源
	http.ServeFile(w, r, objectPath)
}

// get filename from url
func getFileNameFromURL(URL string) (string, error) {
	params := strings.Split(URL, "/")
	fileName := params[len(params)-1]
	fmt.Printf("getFileNameFromURL, len:%d slice=%v\n", len(params), params)
	// url路径过长，返回错误
	if len(params) > PARAMNUMCHECK {
		return "", errors.New("get file name error")
	}
	return fileName, nil
}
func main() {
	// 加载 config.yaml 文件
	if err := settings.Init("./conf/config.yaml"); err != nil {
		fmt.Printf("init config.yaml failed, err:%v\n", err)
		return
	}
	// 获取文件
	http.HandleFunc("/", getHandle)
	// 上传文件
	http.HandleFunc("/upload/", uploadHandler)
	addr := fmt.Sprintf("%s:%d", settings.Conf.Host, settings.Conf.Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err.Error())
	}
}
