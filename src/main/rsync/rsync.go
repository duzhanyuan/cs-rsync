package rsync

import (
	"io/ioutil"
	"fmt"
	"time"
	"datalayer/config"
	"github.com/howeyc/fsnotify"
	"remote"
	"remote/aliyun"
	"sync"
	"local"
	"strings"
	"os"
)

type transferFileSizeTotal struct {
	Size int
	lock sync.Mutex
}

func (t *transferFileSizeTotal) Add(a int) {

	t.lock.Lock()
	t.Size += a
	t.lock.Unlock()
}

func (t *transferFileSizeTotal) Minus(a int) {

	t.lock.Lock()
	t.Size -= a
	t.lock.Unlock()
}

var loadFileSizeTotal = transferFileSizeTotal{0, sync.Mutex{}}
var uploadFileSizeTotal = transferFileSizeTotal{0, sync.Mutex{}}

type SyncOperater struct {
	Info config.SyncInfoer
	ServiceOperater remote.Operater
}

var isQuit = false

var isSync = false

func WaitQuit(ch chan bool) {
	for {
		if isQuit {
			ch <- true
		}
		time.Sleep(time.Duration(10 * 1e9))
	}
}

func Quit() {
	isQuit = true
}

func SyncStart() {
	isSync = true
}

func SyncStop() {
	isSync = false
}

func IsSync() bool {
	return isSync
}

func listenSyncStop(ch chan bool) {
	for {
		if !config.ConfigIsValid() || !isSync {
			ch <- true
		}
		time.Sleep(time.Duration(10 * 1e9))
	}
}

// 启动同步
func Statrt() {
	// 判断是否正在同步
	if !IsSync() {
		SyncStart()

		config.ConfigInit()

		for _, syncInfo := range config.SyncInfoSlice {
			var so SyncOperater
			so.Info = syncInfo
			switch so.Info.Servicer {
			case "aliyun":
				so.ServiceOperater = aliyun.Operater{}
			default:
				so.ServiceOperater = aliyun.Operater{}
			}
			
			so.Info.IsDelete = false

			if so.Info.IsLoad {
				so.LoadBucket()
			} 
			if so.Info.IsUpload {
				so.PutDir()
			}
		}

		for {
			config.ConfigInit()

			n := len(config.SyncInfoSlice)
			
			var chs = make([]chan bool, 2*n)

			for i, syncInfo := range config.SyncInfoSlice {
				var so SyncOperater
				so.Info = syncInfo
				switch so.Info.Servicer {
				case "aliyun":
					so.ServiceOperater = aliyun.Operater{}
				default:
					so.ServiceOperater = aliyun.Operater{}
				}
				
				chs[2*i] = make(chan bool, 1)
				chs[2*i+1] = make(chan bool, 1)

				if so.Info.IsLoad {
					go so.RepetitionLoad(chs[2*i])
				} else {
					chs[2*i] <- true
				}

				if so.Info.IsUpload {
					go so.RepetitionUpload(chs[2*i+1])
				} else {
					chs[2*i+1] <- true
				}
			}
			for _, ch := range chs {
				<-ch
			}

			// 停止同步
			if !IsSync() {
				break
			}

		// 重启，读取新的配置文件
		}
	}	
}

func Stop() {
	SyncStop()
}

func (so SyncOperater) RepetitionLoad(parentCh chan bool) error  {

	var ch = make(chan int, 1)
	so.repetitionLoadSub(ch)
    
	loadUpdateTime := so.Info.LoadUpdateTime
	if loadUpdateTime <= 0 {
		loadUpdateTime = config.LOAD_UPDATE_TIME
	}

    // todo
	// loadUpdateTime = 1
	UpdateTimeCh := make(chan bool, 1)

	for {
		go func() {
			time.Sleep(time.Duration(loadUpdateTime * 1e9))
			UpdateTimeCh <- true
		}()

		stopCh := make(chan bool, 1)
		go listenSyncStop(stopCh)
		select {
			case <- stopCh:
				goto STOP
			case <- ch:
				<- UpdateTimeCh
				so.repetitionLoadSub(ch)
		}
	}
	STOP:

	return nil
}

func (so SyncOperater) repetitionLoadSub(ch chan int) error {
	
	so.LoadBucket()
	ch <- 1
	return nil
}

func (so SyncOperater) LoadBucket() error {
	
	var bucket string
	n := strings.Index(so.Info.RemoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(so.Info.RemoteDir)[:n])
	} else {
		bucket = so.Info.RemoteDir
	}

	filenames, err := local.FindFile(so.Info.LocalDir)
	if err != nil {
		fmt.Println("LoadBucket-FindFile:", err)
		return err
	}

	data, err := so.ServiceOperater.GetObjectOperater().ListAll(so.Info.RemoteDir)
	
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
				MD5Str, _ = local.GetFileMD5(so.Info.LocalDir, filename)
				MD5Str = "\"" + MD5Str + "\""

				if object.ETag == MD5Str {
					objectList = append(objectList[:i], objectList[i+1:]...)
				}

				break
			}

		}

		if !token && so.Info.IsDelete {
			os.Remove(so.Info.LocalDir + "/" + filename)
			local.DelFileMD5(so.Info.LocalDir, filename)
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
			go so.loadBucketSub(object, chs[i])

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
						go so.loadBucketSub(object, chs[j])
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

func (so SyncOperater) loadBucketSub(object remote.Objecter, ch chan int) {
	for {
		if loadFileSizeTotal.Size < config.LOAD_FILES_SIZE {
			loadFileSizeTotal.Add(object.Size)
			break
		}
		time.Sleep(10 * 1e9)
	}

	so.ServiceOperater.GetObjectOperater().Get(so.Info.RemoteDir, so.Info.LocalDir, object.Key)
	ch <- 1
	loadFileSizeTotal.Minus(object.Size)
}

func (so SyncOperater) RepetitionUpload(ch chan bool) error {
	// todo
	
	go so.PutDir()
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		stopCh := make(chan bool, 1)
		go listenSyncStop(stopCh)
		
		for {
			select {
			case <- stopCh:
				goto STOP
			case <-watcher.Event:
				go so.PutDir()
			case err := <-watcher.Error:
				fmt.Println("error:", err)
			}
		}
		STOP:
	}()

	err = watcher.Watch(so.Info.LocalDir)

	fileInfoArr, err := ioutil.ReadDir(so.Info.LocalDir)

	if err != nil {
		ch <- true
		return err
	}

	for _, fileInfo := range fileInfoArr {

		if fileInfo.IsDir() {
			err = watcher.Watch(so.Info.LocalDir + "/" + fileInfo.Name())
		}
	}

	stopCh := make(chan bool, 1)
	go listenSyncStop(stopCh)
	select {
	case <- stopCh:
		goto STOP
	}

	<- done

	/* ... do stuff ... */
	watcher.Close()

	STOP:
	ch <- true
	return nil
}

func (so SyncOperater) PutDir() error {

	var prefix string
	n := strings.Index(so.Info.RemoteDir, "/")    
	if n > 0 {  
		prefix = string([]byte(so.Info.RemoteDir)[n+1:])
	} else {
		prefix = ""
	}

	filenames, err := local.FindFile(so.Info.LocalDir)
	if err != nil {
		fmt.Println("PutDir-FindFile", err)
		return err
	}

	data, err := so.ServiceOperater.GetObjectOperater().ListAll(so.Info.RemoteDir)

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
				MD5Str, _ = local.GetFileMD5(so.Info.LocalDir, filename)
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
				go so.putDirSub(filename, chs[i])
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
							go so.putDirSub(filename, chs[j])
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

	if so.Info.IsDelete {
		for _, object := range objectList {
			so.ServiceOperater.GetObjectOperater().Delete(so.Info.RemoteDir, object.Key)
			local.DelFileMD5(so.Info.LocalDir, object.Key)
		}
	}

	return nil
}

func (so SyncOperater) putDirSub(filename string, ch chan int) {
	so.PutFile(filename)
	ch <- 1
}

func (so SyncOperater) PutFile(filename string) error {
	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(so.Info.LocalDir + "/" + filename)
	if err != nil {
		fmt.Println("PutFile-os.Stat:", err)
		return err
	}

	for {

		if uploadFileSizeTotal.Size < config.UPLOAD_FILES_SIZE {
			uploadFileSizeTotal.Add(int(fileInfo.Size()))
			break
		}

		time.Sleep(5 * 1e9)

	}

	so.ServiceOperater.GetObjectOperater().Put(so.Info.RemoteDir, so.Info.LocalDir, filename)

	uploadFileSizeTotal.Minus(int(fileInfo.Size()))
	return err
}
