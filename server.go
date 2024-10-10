package main

import (
	"database/sql"
	"net/http"
)

type Song struct {
	Group       string `json:"group"`
	Name        string `json:"name"`
	ReleaseDate string `json:"releasedate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

const dblog = "host=localhost port=5432 user=postgres password=123 dbname=effectivemobile sslmode=disable"

func main() {
	http.HandleFunc("/libdata", GetLibData)
	http.HandleFunc("/song", GetSongText)
	http.HandleFunc("/deletesong", DeleteSong)
	http.HandleFunc("/changesong", ChangeSong)
	http.HandleFunc("/addsong", AddSong)
	http.ListenAndServe("localhost:8080", nil)
}

func GetLibData(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	defer db.Close()
}

func GetSongText(w http.ResponseWriter, r *http.Request) {

}

func DeleteSong(w http.ResponseWriter, r *http.Request) {

}

func ChangeSong(w http.ResponseWriter, r *http.Request) {

}

func AddSong(w http.ResponseWriter, r *http.Request) {

}
