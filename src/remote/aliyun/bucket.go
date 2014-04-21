package aliyun

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"errors"
	"remote"
)


func (b BucketOperater)List() (data remote.ListAllMyBucketLister, returnErr error) {

	params := make(map[string]interface{})
	params["VERB"] = "GET"
	params["CanonicalizedResource"] = "/"

	resp, err := request(params)
	if err != nil {
		fmt.Println(err)
		returnErr = err
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var listAllMyBucketsResult ListAllMyBucketsResult
	err = xml.Unmarshal(buf.Bytes(), &listAllMyBucketsResult)

	if err != nil {
		fmt.Println(err)
		returnErr = err
		return
	}

	data.Owner.ID          = listAllMyBucketsResult.Owner.ID
	data.Owner.DisplayName = listAllMyBucketsResult.Owner.DisplayName
	data.BucketList        = make([]remote.Bucketer, 0)
	for _, bucket := range listAllMyBucketsResult.BucketList.Bucket {
		data.BucketList = append(data.BucketList, remote.Bucketer{
			Location: bucket.Location,
			Name    : bucket.Name,
			CreationDate: bucket.CreationDate,
			})
	}

	returnErr = err
	return
}

func (b BucketOperater) Put(name string) error {
	
	params := make(map[string]interface{})
	params["VERB"]   = "PUT"
	params["bucket"] = name
	params["Content-Type"] = "application/x-www-form-urlencoded"
	params["CanonicalizedResource"] = "/" + name + "/"

	resp, err := request(params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return nil
	} else {
		return errors.New(resp.Status)
	}
}

func (b BucketOperater) Delete(name string) error {
	
	params := make(map[string]interface{})
	params["VERB"]   = "DELETE"
	params["bucket"] = name
	params["Content-Type"] = "application/x-www-form-urlencoded"
	params["CanonicalizedResource"] = "/" + name + "/"

	resp, err := request(params)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return nil
	} else {
		return errors.New(resp.Status)
	}
}

func (Operater) GetBucketOperater() remote.BucketOperater {

	return interface{}(BucketOperater{}).(remote.BucketOperater)
}

func (Operater) GetObjectOperater() remote.ObjectOperater {

	return interface{}(ObjectOperater{}).(remote.ObjectOperater)
}

type Operater struct {}

type BucketOperater struct {}

type ObjectOperater struct {}

// ListAllMyBucketsResult
type ListAllMyBucketsResult struct {
	Owner      Owner
	BucketList BucketLister `xml:"Buckets"`
}

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type BucketLister struct {
	Bucket []Bucketer
}

type Bucketer struct {
	Location     string `xml:"Location"`
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

// end ListAllMyBucketsResult
