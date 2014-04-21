package aliyun

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"mapsort"
	"net/url"
	"net"
	"time"
	"strconv"
)

var regionGateway string
var accessKeyId string
var accessKeySecret string

func init() {
	accessKeyId = "XHXt6eD6DTjHdoLK"
	accessKeySecret = "mYSzLXagkew7DNlONDCCyLuVhMABUy"
	regionGateway = "oss.aliyuncs.com"
}

func request(params map[string]interface{}) (resp *http.Response, err error) {

	urlStr := ossGetUrl(params)
	method := params["VERB"]
	data   := params["data"]
	header := ossHttpHeader(params)

	timeout := 60

	methodStr,_ := method.(string)
	dataByteS,_   := data.([]byte)

	req, err := http.NewRequest(methodStr, urlStr, bytes.NewReader(dataByteS))
	if err != nil {
		fmt.Println("PutSmallFile-http.NewRequest()-err", err)
		return
	}

	for key, val := range header {
		req.Header.Add(key, val)
	}

	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				deadline := time.Now().Add(600 * time.Second)
				c, err := net.DialTimeout(netw, addr, time.Second*time.Duration(timeout))
				if err != nil {
					return nil, err
				}
				c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

	resp, err = client.Do(req)

	if err != nil {
		fmt.Println("PutSmallFile-client.Do()", err)
		return
	}

	return
}

func ossHttpHeader(params map[string]interface{}) map[string]string {

	returnVar := make(map[string]string)

	if bucket, ok := params["bucket"]; ok {
		if str, ok := bucket.(string); ok {
			returnVar["Host"] = str + "." + regionGateway
		} else {
			returnVar["Host"] = regionGateway
		}
	} else {
		returnVar["Host"] = regionGateway
	}

	returnVar["Date"] = dateFormatHttp(0)

	if contentType, ok := params["Content-Type"]; ok {
		if str, ok := contentType.(string); ok {
			returnVar["Content-Type"] = str
		} else {
			returnVar["Content-Type"] = ""
		}
	} else {
		returnVar["Content-Type"] = ""
	}

	if contentLength, ok := params["Content-Length"]; ok {
		if str, ok := contentLength.(string); ok {
			returnVar["Content-Length"] = str
		}
	}

	if contentMD5, ok := params["Content-MD5"]; ok {
		if str, ok := contentMD5.(string); ok {
			returnVar["Content-MD5"] = str
		}
	}

	if canonicalizedOSSHeaders, ok := params["CanonicalizedOSSHeaders"]; ok {
		if mapStrIface, ok := canonicalizedOSSHeaders.(map[string]interface{}); ok {
			stringInterfaceSort := mapsort.Sort(mapStrIface)

			for _, item := range stringInterfaceSort {
				if str, ok := item.Val.(string); ok {
					returnVar[item.Key] = str
				}
			}
		}

	}

	returnVar["Authorization"] = "OSS " + accessKeyId + ":" + ossGetSignature(params)

	if headers, ok := params["Headers"]; ok {
		if mapStrStr, ok := headers.(map[string]string); ok {
			for key, val := range mapStrStr {
				returnVar[key] = val
			}
		}
	}

	return returnVar
}

func ossGetSignature(params map[string]interface{}) string {

	dataStrings := ""

	if str, ok := params["VERB"].(string); ok {
		dataStrings += str
	}
	dataStrings += "\n"

	if contentMD5, ok := params["Content-MD5"]; ok {
		if str, ok := contentMD5.(string); ok {
			dataStrings += str
		}
	}
	dataStrings += "\n"

	if contentType, ok := params["Content-Type"]; ok {
		if str, ok := contentType.(string); ok {
			dataStrings += str
		}
	}
	dataStrings += "\n"

	dataStrings += dateFormatHttp(0)
	dataStrings += "\n"

	if OSSHeaders, ok := params["CanonicalizedOSSHeaders"]; ok {
		if mapStrIface, ok := OSSHeaders.(map[string]interface{}); ok {
			stringInterfaceSort := mapsort.Sort(mapStrIface)

			for _, item := range stringInterfaceSort {
				if str, ok := item.Val.(string); ok {
					dataStrings += fmt.Sprintf("%s:%s\n", item.Key, str)
				}
			}
		}

	}

	if resource, ok := params["CanonicalizedResource"]; ok {
		if str, ok := resource.(string); ok {
			dataStrings += str
		}
	}
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	h.Write([]byte(dataStrings))

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func ossGetUrl(params map[string]interface{}) string {
	
	urlStr := "http://"
	if bucket, ok := params["bucket"]; ok {
		if str, ok := bucket.(string); ok {
			urlStr += str + "."
		}
	}

	urlStr += regionGateway + "/"

	if object, ok := params["object"]; ok {
		if str, ok := object.(string); ok {
			urlStr += url.QueryEscape(str)
		}
	}
	if query, ok := params["query"]; ok {
		if mapStrStr, ok := query.(map[string]string); ok {
			queryStr := ""
			for key, val := range mapStrStr {
				if queryStr != "" {
					queryStr += "&"
				}
				queryStr += key + "=" + val
			}
			urlStr += "?" + queryStr
		}
	}

	return urlStr
}

func getSignature(key, content string) string {

	var keyArr = []byte(key)
	var contentArr = []byte(content)
	var h = hmac.New(sha1.New, keyArr)
	h.Write(contentArr)

	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func dateFormatHttp(days int) string {

	var d = time.Now().AddDate(0, 0, days).UTC()
	var weekdayArr = [7]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var monthArr = [12]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	var day = d.Day()
	hour, min, sec := d.Clock()
	var dayStr  = strconv.Itoa(day)
	var hourStr = strconv.Itoa(hour)
	var minStr  = strconv.Itoa(min)
	var secStr  = strconv.Itoa(sec)

	if day < 10 {
		dayStr = "0" + dayStr
	}

	if hour < 10 {
		hourStr = "0" + hourStr
	}

	if min < 10 {
		minStr = "0" + minStr
	}

	if sec < 10 {
		secStr = "0" + secStr
	}

	var dateFormat = weekdayArr[int(d.Weekday())] + ", "
	dateFormat += dayStr + " "
	dateFormat += monthArr[int(d.Month()-1)] + " "
	dateFormat += strconv.Itoa(d.Year()) + " "
	dateFormat += hourStr + ":"
	dateFormat += minStr + ":"
	dateFormat += secStr + " "
	dateFormat += "GMT"
	return dateFormat
}