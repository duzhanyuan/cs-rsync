package main

import (
	"fmt"
	"html/template"
	"io"
	"main/handle"
	"main/rsync"
	"net/http"
	"runtime"
)

func safeHandle(fn func(http.ResponseWriter, *http.Request) io.ReadWriter) (returnHandle http.HandlerFunc) {
	returnHandle = func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			panicVar := recover()
			if err, ok := panicVar.(error); ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tFoot, err := template.ParseFiles("../views/foot.tpl")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tFoot.Execute(w, nil)
		}()

		tHead, err := template.ParseFiles("../views/head.tpl")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session := handle.GlobalSessions.SessionStart(w, r)
		username := session.Get("Username")
		locals := make(map[string]interface{})
		var tBody io.ReadWriter
		// if username == nil {
		tBody = handle.Login(w, r)
		// } else {
		locals["Username"] = username
		tBody = fn(w, r)
		// }
		tHead.Execute(w, locals)
		io.Copy(w, tBody)
	}

	return
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	ch := make(chan bool, 1)
	go rsync.WaitQuit(ch)

	go rsync.Statrt()

	go func() {
		http.HandleFunc("/public/", handle.StaticDir)
		http.HandleFunc("/Login", safeHandle(handle.Login))
		http.HandleFunc("/Logout", safeHandle(handle.Logout))
		http.HandleFunc("/UpdatePassword", safeHandle(handle.UpdatePassword))
		http.HandleFunc("/SyncStart", safeHandle(handle.SyncStart))
		http.HandleFunc("/SyncStop", safeHandle(handle.SyncStop))
		http.HandleFunc("/Quit", safeHandle(handle.Quit))
		http.HandleFunc("/RemoteBucket", safeHandle(handle.RemoteBucket))
		http.HandleFunc("/LocalDir", safeHandle(handle.LocalDir))
		http.HandleFunc("/SyncInfo", safeHandle(handle.SyncInfo))
		http.HandleFunc("/", safeHandle(handle.Index))
		err := http.ListenAndServe(":8080", nil)

		if err != nil {
			fmt.Println("ListenAndServe: ", err.Error())
		}
	}()
	<-ch
}
