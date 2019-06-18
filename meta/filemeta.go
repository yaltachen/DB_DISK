package meta

import (
	"sort"
	"strings"

	mydb "github.com/yaltachen/BD_Disk/db"
)

type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

// 记录所有的 文件元信息
var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)
}

// UpdateFileMeta： 新增/修改文件元信息
func UpdateFileMeta(fileMeta FileMeta) {
	fileMetas[fileMeta.FileSha1] = fileMeta
}

// UpdateFileMetaDB: 新增/修改文件元信息 保存至 MYSQL
func UpdateFileMetaDB(fileMeta FileMeta) bool {
	return mydb.OnFileUploadFinished(fileMeta.FileSha1, fileMeta.FileName,
		fileMeta.FileSize, fileMeta.Location)
}

// GetFileMeta： 获取文件元信息
func GetFileMeta(fileSha1 string) FileMeta {
	return fileMetas[strings.ToLower(fileSha1)]
}

// GetFileMetaDB： 从Mysql获取文件元信息
func GetFileMetaDB(fileSha1 string) (*FileMeta, error) {
	var (
		err   error
		tfile *mydb.TableFile
	)
	if tfile, err = mydb.GetFileMeta(fileSha1); err != nil {
		return nil, err
	}
	return &FileMeta{
		FileName: tfile.FileName.String,
		FileSha1: tfile.FileHash,
		FileSize: tfile.FileSize.Int64,
		Location: tfile.FileAddr.String,
	}, nil
}

// GetLastFileMetas: 获取批量文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	var (
		fMetas []FileMeta
	)
	fMetas = make([]FileMeta, len(fileMetas))

	for _, fMeta := range fileMetas {
		fMetas = append(fMetas, fMeta)
	}

	sort.Sort(ByUploadTime(fMetas))
	return fMetas[0:count]
}

// Remove: 删除文件元信息
func RemoveFileMeta(fsha string) {
	delete(fileMetas, strings.ToLower(fsha))
}
