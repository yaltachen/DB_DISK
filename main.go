package main

import (
	"fmt"
	"net/http"

	"github.com/yaltachen/BD_Disk/handler"
)

func main() {

	var (
		err error
	)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// file 相关接口
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSuc)
	http.HandleFunc("/file/meta", handler.GetFileMetaHander)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileUpdateHandler)
	http.HandleFunc("/file/remove", handler.FileRemoveHandler)
	http.HandleFunc("/file/fastupload", handler.HTTPInterceptor(
		handler.TryFastUploadHandler))

	// 分块上传
	http.HandleFunc("/file/mpupload/init", handler.HTTPInterceptor(
		handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart", handler.HTTPInterceptor(
		handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete", handler.HTTPInterceptor(
		handler.CompleteUploadHandler))

	// user 相关接口
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SigninHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(
		handler.UserInfoHandler))

	if err = http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Fialed to start server, error:%s\r\n", err.Error())
	}
}
