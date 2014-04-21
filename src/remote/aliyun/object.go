package aliyun

import (
	"bytes"
	"datalayer/config"
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	// "net/url"
	"remote"
	"local"
)

func (obj ObjectOperater) List(remoteDir, marker, maxKeys string) (data remote.ObjectLister, returnErr error) {
	var bucket string
	var prefix string
	n := strings.Index(remoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(remoteDir)[:n])
		prefix = string([]byte(remoteDir)[n+1:])
	} else {
		bucket = remoteDir
		prefix = ""
	}

	params := make(map[string]interface{})
	params["VERB"]   = "GET"
	params["bucket"] = bucket
	params["CanonicalizedResource"] = "/" + bucket + "/"

	query := make(map[string]string)
	query["marker"]   = marker
	query["max-keys"] = maxKeys
	query["prefix"]   = prefix
	params["query"]   = query

	resp, err := request(params)

	if err != nil {
		fmt.Println(err)
		returnErr = err
		return
	}

	var buf = new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var listBucketResult ListBucketResult
	err = xml.Unmarshal(buf.Bytes(), &listBucketResult)

	if err != nil {
		fmt.Println(err)
		returnErr = err
		return
	}

	data.Name        = listBucketResult.Name
	data.Prefix      = listBucketResult.Prefix
	data.Marker      = listBucketResult.Marker
	data.MaxKeys     = listBucketResult.MaxKeys
	data.Delimiter   = listBucketResult.Delimiter
	data.IsTruncated = listBucketResult.IsTruncated
	data.ObjectList  = make([]remote.Objecter, 0)
	for _, Object := range listBucketResult.ObjectList {
		data.ObjectList = append(data.ObjectList, remote.Objecter{
			Key: Object.Key,
			LastModified: Object.LastModified,
			ETag: Object.ETag,
			Type: Object.Type,
			Size: Object.Size,
			StorageClass: Object.StorageClass,
			Owner: remote.Owner{ID:Object.Owner.ID,DisplayName:Object.Owner.DisplayName},
			})
	}
	returnErr = err
	return
}

func (obj ObjectOperater) ListAll(remoteDir string) (data remote.ObjectLister, returnErr error) {

	var marker = ""
	var maxKeys = "100"
	data, returnErr = obj.List(remoteDir, marker, maxKeys)

	if returnErr != nil {
		fmt.Println(returnErr)
		return
	}

	var isTruncated = data.IsTruncated
	maxKeys = data.Marker

	for isTruncated {
		dataSub, err := obj.List(remoteDir, marker, maxKeys)

		if err != nil {
			fmt.Println(err)
			return
		}

		isTruncated = dataSub.IsTruncated
		data.ObjectList = append(data.ObjectList, dataSub.ObjectList...)
	}

	return
}

func (obj ObjectOperater) Delete(remoteDir, objectName string) error {
	var bucket string
	var prefix string
	n := strings.Index(remoteDir, "/")    
	if n > 0 {
		bucket = string([]byte(remoteDir)[:n])
		prefix = string([]byte(remoteDir)[n+1:])
	} else {
		bucket = remoteDir
		prefix = ""
	}

	params := make(map[string]interface{})
	params["VERB"]   = "DELETE"
	params["bucket"] = bucket
	params["object"] = prefix + objectName
	params["Date"] = dateFormatHttp(0)

	if config.EncryptToken {
		canonicalizedOSSHeaders := make(map[string]string)
		canonicalizedOSSHeaders["x-oss-server-side-encryption"] = "AES256"
		params["CanonicalizedOSSHeaders"] = canonicalizedOSSHeaders
	}

	params["CanonicalizedResource"] = "/" + bucket + "/" + prefix + objectName
	
	resp, err := request(params)

	if err != nil {
		fmt.Println("DeleteFile-client.Do:", err)
		return err
	}

	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)

	return nil
}

func (obj ObjectOperater) Put(remoteDir, localDir, filename string) error {

	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(localDir + "/" + filename)
	if err != nil {
		fmt.Println("PutFile-os.Stat:", err)
		return err
	}

	if fileInfo.Size() > int64(config.FILE_PART_SIZE) {

		for i := 0; i < config.RETRANSMISSION_NUMBER; i++ {
			err = putLargeObject(remoteDir, localDir, filename)

			if err == nil {
				break
			}
		}
	} else {

		for i := 0; i < config.RETRANSMISSION_NUMBER; i++ {
			err = putSmallObject(remoteDir, localDir, filename)
			if err == nil {
				break
			}
		}
	}

	return err
}

func (obj ObjectOperater) Get(remoteDir, localDir, objectName string) error {
	var bucket string
	var prefix string
	n := strings.Index(remoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(remoteDir)[:n])
		prefix = string([]byte(remoteDir)[n+1:])
	} else {
		bucket = remoteDir
		prefix = ""
	}

	params := make(map[string]interface{})
	params["VERB"] = "GET"
	params["bucket"] = bucket
	params["object"] = prefix + objectName
	
	if config.EncryptToken {
		canonicalizedOSSHeaders := make(map[string]string)
		canonicalizedOSSHeaders["x-oss-server-side-encryption"] = "AES256"
		params["CanonicalizedOSSHeaders"] = canonicalizedOSSHeaders
	}

	params["CanonicalizedResource"] = "/" + bucket + "/" + prefix + objectName

	resp, err := request(params)
	if err != nil {
		fmt.Println("remote-request-fail", err)
		return err
	}	

	objectName = strings.Replace(objectName, prefix, "", 1)
	t, err := os.Create(localDir + "/" + objectName)

	if err != nil {
		var n = strings.Index(objectName, "/")

		if n == -1 {
			fmt.Println("create-file-fail", err)
			return err
		}

		var dirStr = string([]byte(objectName)[0:n])
		err = local.CreateDir(localDir + "/" + dirStr)

		if err != nil {
			fmt.Println("create-dir-fail", err)
			return err
		}

		t, err = os.Create(localDir + "/" + objectName)

		if err != nil {
			fmt.Println("create-file-fail", err)
			return err
		}

		defer t.Close()
	}

	defer t.Close()

	if _, err := io.Copy(t, resp.Body); err != nil {
		fmt.Println("LoadObject-io.Copy()--error", err)
		return err
	}

	defer resp.Body.Close()
	return nil
}

func putSmallObject(remoteDir, localDir, filename string) error {

	fileInfo, err := os.Stat(localDir + "/" + filename)
	if err != nil {
		fmt.Println("putSmallObject-os.Stat: ", err)
		return err
	}

	buf := make([]byte, 0)
	if !fileInfo.IsDir() {
		buf, err = ioutil.ReadFile(localDir + "/" + filename)

		if err != nil {
			fmt.Println("PutSmallFile-ReadFile:", err)
			return err
		}

	}

	var bucket string
	var prefix string
	n := strings.Index(remoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(remoteDir)[:n])
		prefix = string([]byte(remoteDir)[n+1:])
	} else {
		bucket = remoteDir
		prefix = ""
	}

	params := make(map[string]interface{})
	params["VERB"]         = "PUT"
	params["bucket"]       = bucket
	params["object"]       = prefix + filename
	params["Content-Type"] = "application/octet-stream"
	params["data"]         = buf
	if config.EncryptToken {
		canonicalizedOSSHeaders := make(map[string]string)
		canonicalizedOSSHeaders["x-oss-server-side-encryption"] = "AES256"
		params["CanonicalizedOSSHeaders"] = canonicalizedOSSHeaders
	}

	params["CanonicalizedResource"] = "/" + bucket + "/" + prefix + filename

	resp, err := request(params)

	if err != nil {
		fmt.Println("PutSmallFile-client.Do:", err)
		return err
	}

	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)
	return nil
}

func putLargeObject(remoteDir, localDir, filename string) error {
	var bucket string
	var prefix string
	n := strings.Index(remoteDir, "/")    
	if n > 0 {  
		bucket = string([]byte(remoteDir)[:n])
		prefix = string([]byte(remoteDir)[n+1:])
	} else {
		bucket = remoteDir
		prefix = ""
	}

	var fileInfo os.FileInfo
	var err error = nil
	fileInfo, err = os.Stat(localDir + "/" + filename)

	if err != nil {
		return err
	}

	params := make(map[string]interface{})
	params["VERB"]         = "POST"
	params["bucket"]       = bucket
	params["object"]       = prefix + filename
	params["Content-Type"] = "application/octet-stream"
	params["Date"]         = dateFormatHttp(0)

	if config.EncryptToken {
		canonicalizedOSSHeaders := make(map[string]string)
		canonicalizedOSSHeaders["x-oss-server-side-encryption"] = "AES256"
		params["CanonicalizedOSSHeaders"] = canonicalizedOSSHeaders
	}

	CanonicalizedResource := "/" + bucket + "/" + prefix + filename
	CanonicalizedResource += "?uploads"
	params["CanonicalizedResource"] = CanonicalizedResource

	resp, err := request(params)

	if err != nil {
		fmt.Println("PutLargeFile-client.Do", err)
		return err
	}

	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	var initiateMultipartUploadResult InitiateMultipartUploadResult
	err = xml.Unmarshal(buf.Bytes(), &initiateMultipartUploadResult)

	if err != nil {
		fmt.Println("PutLargeFile-xml.Unmarshal", err)
		return err
	}

	uploadId := initiateMultipartUploadResult.UploadId
	bufByte  := make([]byte, config.FILE_PART_SIZE)
	var completeMultipartUpload CompleteMultipartUpload

	for i := 0; i <= int(fileInfo.Size()/int64(config.FILE_PART_SIZE)); i++ {

		file, err := os.Open(localDir + "/" + filename)

		if err != nil {
			fmt.Println("PutLargeFile-os.Open:", err)
			return err
		}

		defer file.Close()

		n, err := file.ReadAt(bufByte, int64(i*config.FILE_PART_SIZE))
		bufByteNew := bufByte[:n]

		params["VERB"] = "PUT"

		CanonicalizedResource = "/" + bucket + "/" + prefix + filename
		CanonicalizedResource += "?partNumber=" + strconv.Itoa(i+1)
		CanonicalizedResource += "&uploadId=" + uploadId

		params["CanonicalizedResource"] = CanonicalizedResource

		query := make(map[string]string)
		query["partNumber"] = strconv.Itoa(i+1)
		query["uploadId"]   = uploadId
		params["query"] = query

		params["data"] = bufByteNew

		resp, err = request(params)

		if err != nil {
			fmt.Println("PutLargeFile-client.Do()-err", err)
			return err
		}

		md5h := md5.New()
		io.Copy(md5h, bytes.NewReader([]byte(bufByteNew)))
		MD5Str := fmt.Sprintf("%X", md5h.Sum(nil))
		MD5Str = "\"" + MD5Str + "\""

		if resp.Header["Etag"][0] != MD5Str {
			fmt.Println("MD5-disaffinity-err")
			return errors.New("MD5-disaffinity")
		}

		completeMultipartUpload.Part = append(completeMultipartUpload.Part, PartS{i + 1, resp.Header["Etag"][0]})
	}

	var completeByte []byte
	completeByte, err = xml.Marshal(&completeMultipartUpload)

	if err != nil {
		fmt.Println("PutLargeFile-xml.Marshal()-err", err)
		return err
	}

	params["VERB"] = "POST"
	params["Date"] = dateFormatHttp(0)
	CanonicalizedResource = "/" + bucket + "/" + prefix + filename
	CanonicalizedResource += "?uploadId=" + uploadId
	params["CanonicalizedResource"] = CanonicalizedResource

	query := make(map[string]string)
	query["uploadId"] = uploadId
	params["query"] = query
	
	params["data"] = completeByte

	_, err = request(params)

	if err != nil {

		fmt.Println("PutLargeFile-client.Do()-err", err)

		return err

	}

	return nil
}

// ListBucketResult
type ListBucketResult struct {
	Name string `xml:"Name"`
	Prefix string `xml:"Prefix"`
	Marker string `xml:"Marker"`
	MaxKeys int `xml:"MaxKeys"`
	Delimiter string `xml:"Delimiter"`
	IsTruncated bool `xml:"IsTruncated"`
	ObjectList []Objecter `xml:"Contents"`
}

type Objecter struct {
	Key string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag string `xml:"ETag"`
	Type string `xml:"Type"`
	Size int `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
	Owner Owner 
}

// end ListBucketResult


//InitiateMultipartUploadResult
type InitiateMultipartUploadResult struct {
	Bucket string `xml:"Bucket"`
	Key string `xml:"Key"`
	UploadId string `xml:"UploadId"`
}
//end InitiateMultipartUploadResult

//CompleteMultipartUpload
type CompleteMultipartUpload struct {
	Part []PartS 
}

type PartS struct {
	PartNumber int `xml:"PartNumber"`
	ETag string `xml:"ETag"`
}
//end CompleteMultipartUpload