package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"io"
)

var FILE_PART_SIZE          int = 5242880  // 1024*1204*5 分片上传文件块大小
var RETRANSMISSION_NUMBER   int = 5        // 出错重传次数
var UPLOAD_GOROUTINE_NUMBER int = 10       // 上传并发协程数量
var UPLOAD_FILES_SIZE       int = 5242880  // 并发上传文件大小上限
var LOAD_GOROUTINE_NUMBER   int = 10       // 下载并发协程数量
var LOAD_FILES_SIZE         int = 5242880  // 并发下载文件大小上限
var LOAD_UPDATE_TIME        int = 6        // 自动更新下载时间
var EncryptToken            bool= false

var configIsValid bool = false             // 配置是否有效

var Config Configer

var SyncInfoSlice SyncInfoSlicer

func init() {
	ConfigInit()
}

func ConfigInit() (err error) {

	configFilePath  := "../data/config.json"
	syncInfoFilePath := "../data/sync_info.json"

	file, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Println("ConfigInit-ioutil.ReadFile:", err)
		return
	}

	err = json.Unmarshal(file, &Config)
	if err != nil {
		fmt.Println("ConfigInit-json.Unmarshal-Config:", err)
		return
	}

	file, err = ioutil.ReadFile(syncInfoFilePath)
	if err != nil {
		fmt.Println("ConfigInit-ioutil.ReadFile:", err)
		return
	}

	err = json.Unmarshal(file, &SyncInfoSlice)
	if err != nil {
		fmt.Println("ConfigInit-json.Unmarshal-SyncInfoSlice:", err)
		return
	}
	configIsValid = true
	return nil
}

func ConfigUpdate() {
	configIsValid = false
	SyncInfoSlice.writeSyncInfo()
}

func ConfigIsValid() bool {
	return configIsValid
}

func (ss *SyncInfoSlicer) Add(s SyncInfoer) *SyncInfoSlicer {
	*ss = append(*ss, s)
	ConfigUpdate()
	return ss
}

func (ss *SyncInfoSlicer) Del(s SyncInfoer) *SyncInfoSlicer {
	for i, v := range *ss {
		if v.RemoteDir  == s.RemoteDir && 
		v.LocalDir  == s.LocalDir  &&
		v.IsLoad    == s.IsLoad    &&
		v.IsUpload  == s.IsUpload  &&
		v.IsDelete  == s.IsDelete  &&
		v.LoadUpdateTime == s.LoadUpdateTime {
			switch i {
			case 0: *ss = (*ss)[1:]
			case len(*ss)-1: *ss = (*ss)[:len(*ss)-1]
			default: *ss = append((*ss)[:i], (*ss)[i+1:]...)
			}
		}
	}
	ConfigUpdate()
	return ss
}

func (ss *SyncInfoSlicer) writeSyncInfo() error {

	dir := "../data/sync_info.json"
	fmt.Println(dir)
	file, err := os.OpenFile(dir, os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err)
		file, err = os.Create(dir)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	defer func() {
		err = file.Close()
		if err != nil {
			fmt.Println(err)
		}
	}()

	jsonByte, err := json.MarshalIndent(ss, "", "	")
	jsonStr := string(jsonByte)
	file.Truncate(int64(len(jsonStr)))

	_, err = io.WriteString(file, jsonStr)

	if err != nil {
		fmt.Println("writeSyncInfo-ioutil.ReadFile:", err)
	}

	if err != nil {
		fmt.Println("writeSyncInfo-json.Marshal:", err)
		return err
	}

	return nil
}

type SyncInfoSlicer []SyncInfoer

type Configer struct {
	FilePartSize          int
	RetransmissionNumber  int
	UploadGoroutineNumber int
	UploadFilesSize       int
	LoadGoroutineNumber   int
	LoadFilesSize         int
	EncryptToken          bool
	KeyId                 string
	KeySecret             string
}

type SyncInfoer struct {
	Servicer  string
	RemoteDir string
	LocalDir  string
	IsLoad    bool
	IsUpload  bool
	IsDelete  bool
	LoadUpdateTime int
}