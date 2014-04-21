package handle

import (
	"local"
	"html/template"
	"net/http"
	"datalayer/config"
	"remote"
	"strconv"
	"main/rsync"
	"remote/aliyun"
	"io"
	"bytes"
)

func Index(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	re = new(bytes.Buffer)
	t, err := template.ParseFiles("../views/index.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	locals := make(map[string]interface{})
	locals["a"] = ""
	t.Execute(re, locals)
	return
}

func SyncStart(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	go rsync.Statrt()
	re = new(bytes.Buffer)
	t, err := template.ParseFiles("../views/sync_start.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(re, nil)
	return
}

func SyncStop(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){

	go rsync.Stop()
	re = new(bytes.Buffer)
	t, err := template.ParseFiles("../views/sync_stop.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	t.Execute(re, nil)
	return
}

func Quit(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	rsync.Quit()
	re = Index(w, r)
	return
}

func SyncInfo(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	re = new(bytes.Buffer)
	r.ParseForm()
	remoteDir := r.FormValue("RemoteDir")
	localDir  := r.FormValue("LocalDir")
	isLoad := false
	if r.FormValue("IsLoad") == "true" {
		isLoad = true
	}
	isUpload := false
	if r.FormValue("IsUpload") == "true" {
		isUpload  = true
	}
	isDelete := false
	if r.FormValue("IsDelete") == "true" {
		isDelete = true
	}
	loadUpdateTime := 0
	if i, _ := strconv.Atoi(r.FormValue("LoadUpdateTime")); i > 0 {
		loadUpdateTime = i
	}
	syncInfo := config.SyncInfoer{
		RemoteDir: remoteDir,
		LocalDir : localDir,
		IsLoad   : isLoad,
		IsUpload : isUpload,
		IsDelete : isDelete,
		LoadUpdateTime : loadUpdateTime,
	}
	operate := r.FormValue("Operate")
	switch operate {
	case "Add":
		config.SyncInfoSlice.Add(syncInfo)
	case "Delete":
		config.SyncInfoSlice.Del(syncInfo)
	}
	
	syncInfoSlice := config.SyncInfoSlice
	
	t, err := template.ParseFiles("../views/sync_info.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	locals := make(map[string]interface{})
	locals["SyncInfoSlice"] = syncInfoSlice
	
	t.Execute(re, locals)
	return
}

func RemoteBucket(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	re = new(bytes.Buffer)
	r.ParseForm()
	servicer   := r.FormValue("Servicer")
	bucketName := r.FormValue("BucketName")
	var serviceOperater remote.Operater
	switch servicer {
	case "aliyun":
		serviceOperater = aliyun.Operater{}
	default:
		serviceOperater = aliyun.Operater{}
	}
	operate    := r.FormValue("Operate")
	switch operate {
	case "Add":
		serviceOperater.GetBucketOperater().Put(bucketName)
	case "Delete":
		serviceOperater.GetBucketOperater().Delete(bucketName)
	}
	
	bucketList, _ := serviceOperater.GetBucketOperater().List()
	locals := make(map[string]interface{})
	locals["BucketList"] = bucketList.BucketList

	t, err := template.ParseFiles("../views/remote_bucket.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(re, locals)
	return
}

func LocalDir(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){

	re = new(bytes.Buffer)
	t, err := template.ParseFiles("../views/bucket.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	locals := make(map[string]interface{})
	t.Execute(re, locals)
	return
}

func StaticDir(w http.ResponseWriter, r *http.Request) {
	file := "../views/public" + r.URL.Path[len("/public"):]
	if exists := local.IsFileExist(file); !exists {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, file)
	return
}