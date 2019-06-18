package handler

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	rPool "github.com/yaltachen/BD_Disk/cache/redis"
	dblayer "github.com/yaltachen/BD_Disk/db"
	"github.com/yaltachen/BD_Disk/util"
)

// MultipartUploadInfo： 分块上传的初始化信息
type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		username   string
		filehash   string
		filesize   int
		rConn      redis.Conn
		uploadinfo MultipartUploadInfo
	)
	// 1.解析用户请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	filehash = r.Form.Get("filehash")
	filesize, _ = strconv.Atoi(r.Form.Get("filesize"))
	// 2.获得redis的一个连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()
	// 3.生成分块上传的初始化信息
	uploadinfo = MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, // 5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}
	// 4.将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+uploadinfo.UploadID, "chunkcount", uploadinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+uploadinfo.UploadID, "filehash", uploadinfo.FileHash)
	rConn.Do("HSET", "MP_"+uploadinfo.UploadID, "filesize", uploadinfo.FileSize)

	// 5.将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", uploadinfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//TODO: 客户端 + 服务端校验每个分块的hash值是否相同
	var (
		// username string
		uploadID string
		chunkIdx string
		rConn    redis.Conn // redis连接
		fPath    string     // 文件路径
		fd       *os.File   // 文件句柄
		err      error
		buf      []byte
		n        int // read bytes
	)
	// 1.解析用户参数
	r.ParseForm()
	// username = r.Form.Get("username")
	uploadID = r.Form.Get("uploadid")
	chunkIdx = r.Form.Get("index")
	// // 2.获取redis连接池中的一个连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()
	// 3.获取文件句柄，用于存储分块信息
	fPath = "./data/" + uploadID + "/" + chunkIdx
	os.MkdirAll(path.Dir(fPath), 0744)
	if fd, err = os.Create(fPath); err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf = make([]byte, 1024*1024) // 每次读1MB
	for {
		n, err = r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 4.更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIdx, 1)
	// 5.处理结果返回客户端
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	var (
		uploadID   string
		username   string
		filehash   string
		filesize   int64
		filename   string
		rConn      redis.Conn
		data       []interface{}
		err        error
		totalCount int
		chunkCount int
	)
	// 1.解析请求参数
	r.ParseForm()
	username = r.Form.Get("username")
	uploadID = r.Form.Get("uploadid")
	filehash = r.Form.Get("filehash")
	filesize, _ = strconv.ParseInt(r.Form.Get("filesize"), 10, 64)
	filename = r.Form.Get("filename")

	// 2.获得redis连接池中的连接
	rConn = rPool.RedisPool().Get()
	defer rConn.Close()
	// 3.通过uploadid查询是否所有分块上传完成
	if data, err = redis.Values(rConn.Do("HGETALL", "MP_"+uploadID)); err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	for i := 0; i < len(data); i += 2 {
		key := string(data[i].([]byte))
		value := string(data[i+1].([]byte))
		if key == "chunkcount" {
			totalCount, _ = strconv.Atoi(value)
		} else if strings.HasPrefix(key, "chkidx_") && value == "1" {
			chunkCount += 1
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
	}
	// TODO: 4.合并分块
	// 5.更新唯一文件表以及用户文件表

	dblayer.OnFileUploadFinished(filehash, filename, filesize, "TODO()")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, filesize)
	// 6.响应处理结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}
