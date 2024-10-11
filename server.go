package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Song struct {
	Group       string `json:"group"`
	Name        string `json:"name"`
	ReleaseDate string `json:"releasedate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type DataFilter struct {
	Limit       int  `json:"limit"`
	Page        int  `json:"page"`
	ID          bool `json:"id"`
	Group       bool `json:"group"`
	Name        bool `json:"name"`
	ReleaseDate bool `json:"releasedate"`
	Text        bool `json:"text"`
	Link        bool `json:"link"`
}

var dblog string

func init() {
	if err := godotenv.Load("go.env"); err != nil {
		panic(err)
	}
	var exist bool
	dblog, exist = os.LookupEnv("dblog")
	log.Println(exist, dblog)
}
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
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		panic(err)
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		panic(err)
	}
	res, err := db.Query(fmt.Sprintf("select * from songs order by id limit %v offset %v", limit, (page-1)*limit))
	songs := []Song{}
	for res.Next() {
		song := Song{}
		if err = res.Scan(&song); err != nil {
			panic(err)
		}
		songs = append(songs, song)
	}
	for i := range songs {
		fmt.Fprint(w, songs[i])
	}
}

func GetSongText(w http.ResponseWriter, r *http.Request) {

}

func DeleteSong(w http.ResponseWriter, r *http.Request) {

}

func ChangeSong(w http.ResponseWriter, r *http.Request) {

}

func AddSong(w http.ResponseWriter, r *http.Request) {
	var song Song
	err := json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		panic(err)
	}
	fmt.Println(song)
	resp, err := http.Get("http://127.0.0.1:8070/info")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	songFiller := Song{}
	err = json.NewDecoder(resp.Body).Decode(&songFiller)
	if err != nil {
		panic(err)
	}
	song.ReleaseDate = songFiller.ReleaseDate
	song.Text = songFiller.Text
	song.Link = songFiller.Link
	fmt.Println(song)
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	res, err := db.Exec(fmt.Sprintf("insert into songs(\"group\", name, releasedate, text, link) values('%v', '%v', '%v', '%v', '%v')",
		song.Group, song.Name, song.ReleaseDate, song.Text, song.Link))
	if err != nil {
		panic(err)
	}
	log.Println(res)
}
