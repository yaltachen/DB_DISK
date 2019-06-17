package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/yaltachen/emulate/BD_Disk/FILESTORE-SERVER/meta"
	"github.com/yaltachen/emulate/BD_Disk/FILESTORE-SERVER/util"
)

// 处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// 返回上传页面
		getUploadHandler(w, r)
	} else if r.Method == http.MethodPost {
		// 处理文件上传
		postUploadHandler(w, r)
	}
}
func getUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		err  error
		data []byte
	)
	if data, err = ioutil.ReadFile("./static/index.html"); err != nil {
		io.WriteString(w, "internal server error")
		return
	}
	io.WriteString(w, string(data))
}
func postUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 接收文件流以及存储到本地目录
	// 保存文件元信息
	var (
		err        error
		file       multipart.File
		fileHeader *multipart.FileHeader
		fileMeta   meta.FileMeta
		localFile  *os.File
	)
	if file, fileHeader, err = r.FormFile("file"); err != nil {
		fmt.Printf("Failed to get data, err: %s\n", err.Error())
		return
	}
	defer file.Close()

	fileMeta = meta.FileMeta{
		FileName: fileHeader.Filename,
		Location: "./tmp/" + fileHeader.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	fileMeta.FileName = fileHeader.Filename
	if localFile, err = os.Create(fileMeta.Location); err != nil {
		fmt.Printf("Failed to create file, err: %s\n", err.Error())
	}
	defer localFile.Close()

	if fileMeta.FileSize, err = io.Copy(localFile, file); err != nil {
		fmt.Printf("Failed to save data into file, err: %s\n", err.Error())
		return
	}

	localFile.Seek(0, 0)
	fileMeta.FileSha1 = util.FileSha1(localFile)

	// meta.UpdateFileMeta(fileMeta)
	meta.UpdateFileMetaDB(fileMeta)

	http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
}

// 上传成功
func UploadSuc(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload success")
	return
}

// 获取单个文件元信息
func GetFileMetaHander(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var (
		fileHash string
		fMeta    meta.FileMeta
		data     []byte
		err      error
	)

	fileHash = r.Form["filehash"][0]
	// fMeta = meta.GetFileMeta(fileHash)
	if fMeta, err = meta.GetFileMetaDB(fileHash); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if data, err = json.Marshal(fMeta); err != nil {
		fmt.Printf("Failed to convert to json, err: %s\r\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 获取多个文件信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	var (
		count  int
		fMetas []meta.FileMeta
		data   []byte
		err    error
	)

	count, _ = strconv.Atoi(r.Form.Get("limit"))
	fMetas = meta.GetLastFileMetas(count)

	if data, err = json.Marshal(fMetas); err != nil {
		fmt.Printf("Failed to conver to json, err: %v\r\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

// 下载单个文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		fSha1 string
		fMeta meta.FileMeta
		file  *os.File
		err   error
		data  []byte
	)
	r.ParseForm()

	fSha1 = r.Form.Get("filehash")
	fMeta = meta.GetFileMeta(fSha1)

	if file, err = os.Open(fMeta.Location); err != nil {
		fmt.Printf("Failed to open file, err: %v\r\n", err)
		return
	}

	defer file.Close()

	if data, err = ioutil.ReadAll(file); err != nil {
		fmt.Printf("Failed to read file, err: %v\r\n", err)
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-disposition", "attachment;filename=\""+fMeta.FileName+"\"")
	w.Write(data)
}

// 更新元信息（重命名）
func FileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var (
		opType      string
		fileSha1    string
		newFileName string
		curFile     meta.FileMeta
		data        []byte
		err         error
	)
	r.ParseForm()

	opType = r.Form.Get("op")
	fileSha1 = r.Form.Get("filehash")
	newFileName = r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFile = meta.GetFileMeta(fileSha1)
	curFile.FileName = newFileName
	meta.UpdateFileMeta(curFile)

	if data, err = json.Marshal(curFile); err != nil {
		fmt.Printf("Failed to convert to json, err: %v\r\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func FileRemoveHandler(w http.ResponseWriter, r *http.Request) {
	var (
		fileSha1 string
		fileMeta meta.FileMeta
	)
	r.ParseForm()
	fileSha1 = r.Form["filehash"][0]

	// 删除文件
	fileMeta = meta.GetFileMeta(fileSha1)
	os.Remove(fileMeta.Location)

	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
}
