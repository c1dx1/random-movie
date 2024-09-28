package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

func addFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Пользователь не авторизован"})
		return
	}

	movieID, err := strconv.Atoi(r.FormValue("movie_id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Некорректный ID фильма"})
		return
	}
	movieName := r.FormValue("movie_name")

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Пользователь не найден"})
		return
	}

	err = addToFavorites(userID, movieID, movieName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"success": "false", "message": "Ошибка при добавлении в избранное"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "message": "Фильм добавлен в избранное"})
}

func removeFavoriteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)

	if !ok {
		http.Error(w, "Not auth", http.StatusUnauthorized)
		return
	}

	movieID, err := strconv.Atoi(r.FormValue("movie_id"))
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	err = removeFromFavorites(userID, movieID) // Напиши эту функцию, чтобы удалять фильм из базы данных
	if err != nil {
		http.Error(w, "Error removing from favorites", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/favorites", http.StatusSeeOther)
}

func favoriteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	username, ok := session.Values["username"].(string)

	if !ok {
		http.Error(w, "Not auth", http.StatusInternalServerError)
		return
	}

	var userID int
	err := db.QueryRow("SELECT id FROM users WHERE username=$1", username).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
	}

	if r.Method == "GET" {
		movies, err := getFavorites(userID)
		if err != nil {
			http.Error(w, "Error getting favorites", http.StatusInternalServerError)
		}

		tmpl, err := template.ParseFiles("templates/favorites.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, movies)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
