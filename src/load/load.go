package load

import (
	"datalayer/config"
	"fmt"
	"os"
	"remote"
	"strings"
	"sync"
	"time"
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

func LoadBucket(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer) error {
	
	var bucket string
	n := strings.Index(syncInfo.RemoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(syncInfo.RemoteDir)[:n])
	} else {
		bucket = syncInfo.RemoteDir
	}

	filenames, err := local.FindFile(syncInfo.LocalDir)
	if err != nil {
		fmt.Println("LoadBucket-FindFile:", err)
		return err
	}

	data, err := bucketOperater.GetObjectOperater().ListAll(syncInfo.RemoteDir)
	
	if err != nil {
		fmt.Println("LoadBucket-remote.AllObjectList:", err)
		return err
	}

	var objectList = data.ObjectList
	var MD5Str string
	var token bool

	for _, filename := range filenames {
		token = false

		for i, object := range objectList {

			if object.Key == filename {
				token = true
				MD5Str, _ = local.GetFileMD5(syncInfo.LocalDir, filename)
				MD5Str = "\"" + MD5Str + "\""

				if object.ETag == MD5Str {
					objectList = append(objectList[:i], objectList[i+1:]...)
				}

				break
			}

		}

		if !token && syncInfo.IsDelete {
			os.Remove(syncInfo.LocalDir + "/" + filename)
			local.DelFileMD5(syncInfo.LocalDir, filename)
		}

	}

	var chs = make([]chan int, config.LOAD_GOROUTINE_NUMBER)
	var i = 0
	var ii = 0

	for _, object := range objectList {
		fileInfo := local.GetFileInfo(bucket)
		fileInfo[object.Key] = object.ETag
		defer local.WriteFileInfo(bucket)
		if i < config.LOAD_GOROUTINE_NUMBER {
			chs[i] = make(chan int, 1)
			go loadBucketSub(bucketOperater, syncInfo, object, chs[i])

			i++
		} else {
			var timeout = make(chan bool, 1)

			for {
				ii = -1

				for j := 0; j < config.LOAD_GOROUTINE_NUMBER; j++ {

					go func() {

						time.Sleep(time.Duration(100 * 1e9))
						timeout <- true
					}()

					select {
					case <-chs[j]:
						ii = 1
					case <-timeout:
					}
					if ii == 1 {
						go loadBucketSub(bucketOperater, syncInfo, object, chs[j])
						break
					}

				}

				if ii == 1 {
					break
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

	return nil
}

func loadBucketSub(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, object remote.Objecter, ch chan int) {
	for {
		if transferFilesP.Size < config.LOAD_FILES_SIZE {
			transferFilesP.Add(object.Size)
			break
		}
		time.Sleep(10 * 1e9)
	}

	bucketOperater.GetObjectOperater().Get(syncInfo.RemoteDir, syncInfo.LocalDir, object.Key)
	ch <- 1
	transferFilesP.Minus(object.Size)
}
