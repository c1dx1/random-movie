package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
)

var db *sql.DB

func initDB() {
	connStr := os.Getenv("URL_DB")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected!")
}

func addToFavorites(userID, movieID int, movieName string) error {
	_, err := db.Exec("INSERT INTO favorites (user_id, movie_id, movie_name) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING", userID, movieID, movieName)
	return err
}

func getFavorites(userID int) ([]MovieData, error) {
	rows, err := db.Query("SELECT movie_id, movie_name FROM favorites WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []MovieData
	for rows.Next() {
		var movie MovieData
		err := rows.Scan(&movie.ID, &movie.Name)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func removeFromFavorites(userID, movieID int) error {
	_, err := db.Exec("DELETE FROM favorites WHERE user_id=$1 AND movie_id=$2", userID, movieID)
	return err
}

func registerUser(username, password string) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ($1, $2)", username, string(passwordHash))
	if err != nil {
		return err
	}

	return nil
}

func loginUser(username, password string) bool {
	var passwordHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username=$1", username).Scan(&passwordHash)
	if err != nil {
		return false
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	return err == nil
}
