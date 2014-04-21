package upload

import (
	"datalayer/config"
	"fmt"
	"os"
	"sync"
	"time"
	"remote"
	"strings"
	// "net/url"
	"local"
)

var transferFilesSizeLock sync.Mutex

type transferFiles struct {
	Size int
}

func (t *transferFiles) Add(a int) {

	transferFilesSizeLock.Lock()
	t.Size += a
	transferFilesSizeLock.Unlock()
}

func (t *transferFiles) Minus(a int) {

	transferFilesSizeLock.Lock()
	t.Size -= a
	transferFilesSizeLock.Unlock()
}

var transferFilesP = transferFiles{0}

func PutFile(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, filename string) error {
	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(syncInfo.LocalDir + "/" + filename)
	if err != nil {
		fmt.Println("PutFile-os.Stat:", err)
		return err
	}

	for {

		if transferFilesP.Size < config.UPLOAD_FILES_SIZE {
			transferFilesP.Add(int(fileInfo.Size()))
			break
		}

		time.Sleep(5 * 1e9)

	}

	bucketOperater.GetObjectOperater().Put(syncInfo.RemoteDir, syncInfo.LocalDir, filename)

	transferFilesP.Minus(int(fileInfo.Size()))
	return err
}

func PutDir(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer) error {

	var prefix string
	n := strings.Index(syncInfo.RemoteDir, "/")    
	if n > 0 {  
		prefix = string([]byte(syncInfo.RemoteDir)[n+1:])
	} else {
		prefix = ""
	}

	filenames, err := local.FindFile(syncInfo.LocalDir)
	if err != nil {
		fmt.Println("PutDir-FindFile", err)
		return err
	}

	data, err := bucketOperater.GetObjectOperater().ListAll(syncInfo.RemoteDir)

	if err != nil {
		fmt.Println("PutDir-remote.AllObjectList", err)
		return err
	}

	var objectList = data.ObjectList

	var chs = make([]chan int, config.UPLOAD_GOROUTINE_NUMBER)
	var i = 0
	var ii int
	var token bool
	var MD5Str string

	for _, filename := range filenames {

		token = false

		for j, object := range objectList {

			if object.Key == (prefix + filename) {
				token = true
				MD5Str, _ = local.GetFileMD5(syncInfo.LocalDir, filename)
				MD5Str = "\"" + MD5Str + "\""

				if object.ETag != MD5Str {
					token = false
				}

				objectList = append(objectList[:j], objectList[j+1:]...)
				break
			}

		}

		if !token {

			if i < config.UPLOAD_GOROUTINE_NUMBER {
				chs[i] = make(chan int, 1)
				go putDirSub(bucketOperater, syncInfo, filename, chs[i])
				i++

			} else {

				for {
					ii = -1

					for j := 0; j < config.UPLOAD_GOROUTINE_NUMBER; j++ {
						var timeout = make(chan bool, 1)

						go func() {
							time.Sleep(time.Duration(1 * 1e9))
							timeout <- true
						}()

						select {
						case <-chs[j]:
							ii = 1
						case <-timeout:
						}

						if ii == 1 {
							go putDirSub(bucketOperater, syncInfo, filename, chs[j])
							break
						}

					}

					if ii == 1 {
						break
					}
				}
			}

		}

	}

	for j, ch := range chs {

		if j == i {
			break
		}

		<-ch

	}

	if syncInfo.IsDelete {
		for _, object := range objectList {
			bucketOperater.GetObjectOperater().Delete(syncInfo.RemoteDir, object.Key)
			local.DelFileMD5(syncInfo.LocalDir, object.Key)
		}
	}

	return nil
}

func putDirSub(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, filename string, ch chan int) {
	PutFile(bucketOperater, syncInfo, filename)
	ch <- 1
}