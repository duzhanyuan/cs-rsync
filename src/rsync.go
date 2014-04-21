package rsyncaaaa

import (
	"io/ioutil"
	"main/load"
	"main/upload"
	"fmt"
	"time"
	"datalayer/config"
	"github.com/howeyc/fsnotify"
	"remote"
	"remote/aliyun"
)

var isQuit = false

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

// 启动同步
func Statrt() {
	// 判断是否正在同步
	if !config.IsSync() {
		config.SyncStart()

		for {
			config.ConfigInit()

			n := len(config.SyncInfoSlice)
			
			var chs = make([]chan bool, 2*n)

			for i, syncInfo := range config.SyncInfoSlice {

				var bucketOperater remote.BucketOperater
				switch syncInfo.Servicer {
				case "aliyun":
					bucketOperater = aliyun.Bucketer{}
				default:
					bucketOperater = aliyun.Bucketer{}
				}
				
				chs[2*i] = make(chan bool, 1)
				chs[2*i+1] = make(chan bool, 1)

				if syncInfo.IsLoad {
					go RepetitionLoad(bucketOperater, syncInfo, chs[2*i])
				} else {
					chs[2*i] <- true
				}

				if syncInfo.IsUpload {
					go RepetitionUpload(bucketOperater, syncInfo, chs[2*i+1])
				} else {
					chs[2*i+1] <- true
				}
			}
			for _, ch := range chs {
				<-ch
			}

			// 停止同步
			if !config.IsSync() {
				break
			}

		// 重启，读取新的配置文件
		}
	}	
}

func Stop() {
	config.SyncStop()
}

func RepetitionUpload(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, ch chan bool) error {
	// todo
	
	go upload.PutDir(bucketOperater, syncInfo)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		stopCh := make(chan bool, 1)
		go config.ListenSyncStop(stopCh)
		
		for {
			select {
			case <- stopCh:
				goto STOP
			case <-watcher.Event:
				go upload.PutDir(bucketOperater, syncInfo)
			case err := <-watcher.Error:
				fmt.Println("error:", err)
			}
		}
		STOP:
	}()

	err = watcher.Watch(syncInfo.LocalDir)

	fileInfoArr, err := ioutil.ReadDir(syncInfo.LocalDir)

	if err != nil {
		ch <- true
		return err
	}

	for _, fileInfo := range fileInfoArr {

		if fileInfo.IsDir() {
			err = watcher.Watch(syncInfo.LocalDir + "/" + fileInfo.Name())
		}
	}

	stopCh := make(chan bool, 1)
	go config.ListenSyncStop(stopCh)
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

func RepetitionLoad(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, parentCh chan bool) error  {

	var ch = make(chan int, 1)
	repetitionLoadSub(bucketOperater, syncInfo, ch)
    
	loadUpdateTime := syncInfo.LoadUpdateTime
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
		go config.ListenSyncStop(stopCh)
		select {
			case <- stopCh:
				goto STOP
			case <- ch:
				<- UpdateTimeCh
				repetitionLoadSub(bucketOperater, syncInfo, ch)
		}
	}
	STOP:

	return nil
}

func repetitionLoadSub(bucketOperater remote.BucketOperater, syncInfo config.SyncInfoer, ch chan int) error {
	load.LoadBucket(bucketOperater, syncInfo)

	ch <- 1
	return nil
}