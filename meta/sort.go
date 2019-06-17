package meta

import "time"

const baseFormat = "2006-01-02 15:04:05"

type ByUploadTime []FileMeta

func (a ByUploadTime) Len() int {
	return len(a)
}

func (a ByUploadTime) Less(i int, j int) bool {
	iTime, _ := time.Parse(baseFormat, a[i].UploadAt)
	jTime, _ := time.Parse(baseFormat, a[j].UploadAt)
	return iTime.UnixNano() > jTime.UnixNano()
}

func (a ByUploadTime) Swap(i int, j int) {
	a[i], a[j] = a[j], a[i]
}
