package handle

import (
	"main/session"
	"net/http"
	"html/template"
	"datalayer/account"
	"io"
	"bytes"
)


var memoryProvid session.MemoryProvider
var GlobalSessions *session.Manager

func init() {
	memoryProvid.Init()
	session.Register("memory", &memoryProvid)
	GlobalSessions, _ = session.NewManager("memory","GoSessionId",3600)
	go GlobalSessions.GC()
}


func Login(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){

	re = new(bytes.Buffer)
	
	session := GlobalSessions.SessionStart(w, r)
	username := session.Get("Username")
	if username != nil {
		http.Redirect(w, r, "/", 302)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("Username")
		password := r.FormValue("Password")
		if account.CheckPassword(username, password) {
			session.Set("Username", username)
			if r.Referer() == "" {
				http.Redirect(w, r, "/", 302)
			} else {
				http.Redirect(w, r, r.Referer(), 302)
			}
			return
		}
		t, err := template.ParseFiles("../views/login_failure.tpl")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		t.Execute(re, nil)
	}

	t, err := template.ParseFiles("../views/login.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(re, nil)
	return
}

func Logout(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	re = new(bytes.Buffer)
	GlobalSessions.SessionDestroy(w, r)
	http.Redirect(w, r, "/", 302)
	return
}

func UpdatePassword(w http.ResponseWriter, r *http.Request) (re io.ReadWriter){
	
	re = new(bytes.Buffer)
	if r.Method == "POST" {
		r.ParseForm()
		session := GlobalSessions.SessionStart(w, r)
		username := session.Get("Username").(string)
		password := r.FormValue("Password")
		newPassword := r.FormValue("NewPassword")
		newPasswordConfirm := r.FormValue("NewPasswordConfirm")
		if newPassword != newPasswordConfirm {
			t, err := template.ParseFiles("../views/update_password_error_1.tpl")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t.Execute(re, nil)
		}else if !account.CheckPassword(username, password) {
			t, err := template.ParseFiles("../views/update_password_error_2.tpl")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t.Execute(re, nil)
		} else if !account.UpdatePassword(username, password, newPassword) {
			t, err := template.ParseFiles("../views/update_password_error_3.tpl")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t.Execute(re, nil)
		} else {
			if !account.CheckPassword(username, password) {
				t, err := template.ParseFiles("../views/update_password_success.tpl")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				t.Execute(re, nil)
				GlobalSessions.SessionDestroy(w, r)
				var p []byte
				Login(w, r).Read(p)
				re.Write(p)
				return
			}
		}
	} 
	t, err := template.ParseFiles("../views/update_password.tpl")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(re, nil)
	return
}
