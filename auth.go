package main

import (
	_ "github.com/lib/pq"
	"html/template"
	"net/http"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if loginUser(username, password) {
			session.Values["username"] = username
			session.Save(r, w)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			http.Error(w, "Fail: invalid credentials", http.StatusUnauthorized)
		}
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	} else if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		err := registerUser(username, password)
		if err != nil {
			http.Error(w, "Fail: register user", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["username"] = ""
	session.Options.MaxAge = -1

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
