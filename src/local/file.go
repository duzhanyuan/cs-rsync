package local

import (
	"crypto/md5"
	"encoding/json"
	"strings"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"datalayer/config"
)

type FileInfoer map[string]string

var fileInfoMap map[string]FileInfoer

var initToken bool

func init() {
	if !initToken {
		fileInfoMap = make(map[string]FileInfoer)
		initToken = true
	}
	InitFileInfo()
}

func InitFileInfo() {
	config.ConfigInit()
	for _, syncInfo := range config.SyncInfoSlice {
		filenames, _ := FindFile(syncInfo.LocalDir)
		fileInfo := GetFileInfo(syncInfo.LocalDir)

		for key, _ := range fileInfo {
			delete(fileInfo, key)
		}

		for _, filename := range filenames {
			MD5Str, _ := createFileMD5(syncInfo.LocalDir+"/"+filename)
			fileInfo[filename] = MD5Str
		}

		WriteFileInfo(syncInfo.LocalDir)
	}
}

func GetFileInfo(localDir string) FileInfoer {

	if len(localDir) > 0 && localDir[len(localDir)-1] == '/' {
		localDir = localDir[0:len(localDir)-1]
	}
	localDir = strings.Replace(localDir, ":/", "_", -1)
	localDir = strings.Replace(localDir, ":\\", "_", -1)
	localDir = strings.Replace(localDir, "/", "_", -1)
	fileInfo, ok := fileInfoMap[localDir]
	if ok {
		return fileInfo
	} else {
		fileInfoMap[localDir] = FileInfoer{}

		path := fmt.Sprintf("../data/%s_file_info.json", localDir)
		file, err := ioutil.ReadFile(path)
		if err != nil {
			return fileInfoMap[localDir]
		}

		var fileInfoTmp FileInfoer
		fileInfoTmp = make(map[string]string)
		err = json.Unmarshal(file, &fileInfoTmp)
		if err != nil {
			fmt.Println(err)
		}
		fileInfoMap[localDir] = fileInfoTmp
		return fileInfoMap[localDir]
	}
}

func WriteFileInfo(localDir string) error {
	if localDir[len(localDir)-1] == '/' {
		localDir = localDir[0:len(localDir)-1]
	}
	localDir = strings.Replace(localDir, ":/", "_", -1)
	localDir = strings.Replace(localDir, ":\\", "_", -1)
	localDir = strings.Replace(localDir, "/", "_", -1)
	dir := fmt.Sprintf("../data/%s_file_info.json", localDir)
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

	fileInfoTmp, ok := fileInfoMap[localDir]
	if !ok {
		return errors.New("FileInfo:NotExist")
	}
	jsonByte, err := json.MarshalIndent(&fileInfoTmp, "", "	")
	jsonStr := string(jsonByte)
	file.Truncate(int64(len(jsonStr)))

	_, err = io.WriteString(file, jsonStr)

	if err != nil {
		fmt.Println("ConfigInit-ioutil.ReadFile:", err)
	}

	if err != nil {
		fmt.Println("ConfigInit-json.Marshal:", err)
		return err
	}

	return nil
}

func GetFileMD5(localDir, filename string) (MD5 string, returnErr error) {
	fileInfo := GetFileInfo(localDir)
	if val, ok := fileInfo[filename]; ok {
		return val, nil
	}
	MD5, returnErr = createFileMD5(localDir + "/" + filename)
	if returnErr == nil {
		fileInfo[filename] = MD5
		WriteFileInfo(localDir)
	}
	return
}

func createFileMD5(filename string) (MD5 string, returnErr error) {

	file, returnErr := os.Open(filename)
	if returnErr != nil {
		return
	}
	defer file.Close()

	md5h := md5.New()
	io.Copy(md5h, file)
	MD5 = fmt.Sprintf("%X", md5h.Sum(nil))
	return
}

func DelFileMD5(localDir, filename string) {
	fileInfo := GetFileInfo(localDir)
	delete(fileInfo, filename)
	return
}

func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return os.IsExist(err)
}