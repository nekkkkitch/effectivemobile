package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Song struct {
	Id          int
	Group       string `json:"group"`
	Name        string `json:"name"`
	ReleaseDate string `json:"releasedate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type DataFilter struct {
	Limit       int  `json:"limit"`
	Page        int  `json:"page"`
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
	var datafilter DataFilter
	err = json.NewDecoder(r.Body).Decode(&datafilter)
	if err != nil {
		panic(err)
	}
	limit := datafilter.Limit
	page := datafilter.Page
	log.Printf("Songs per page: %v\nCurrent page: %v\n", limit, page)
	res, err := db.Query(fmt.Sprintf("select * from songs order by id limit %v offset %v", limit, (page-1)*limit))
	var songs []Song
	for res.Next() {
		var song Song
		if err = res.Scan(&song.Id, &song.Group, &song.Name, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
			panic(err)
		}
		songs = append(songs, song)
	}
	log.Printf("Songs to show: %v", songs)
	var msg string
	for i := range songs {
		msg += fmt.Sprintf("%v) ", i+1)
		songValues := reflect.ValueOf(songs[i])
		songTypes := songValues.Type()
		filterValues := reflect.ValueOf(datafilter)
		filterTypes := filterValues.Type()
		for i := 0; i < songValues.NumField(); i++ {
			for j := 0; j < filterValues.NumField(); j++ {
				if filterTypes.Field(j).Name == songTypes.Field(i).Name {
					if filterValues.Field(j).Bool() {
						msg += songTypes.Field(i).Type.Name() + ": " + songValues.Field(i).String() + "\n"
					}
				}
			}
		}
		msg += "\n"
	}
	fmt.Fprintln(w, msg)
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

/*
func BuildQuery(requestedData []string, limit, page int) string {
	var query string
	query = fmt.Sprintf("select (%v) from songs order by id limit %v offset %v",
		strings.Join(requestedData, ", "), limit, (page-1)*limit)
	log.Println("Result query: " + query)
	return query
}


func RequestedData(filter DataFilter) (int, int, []string) {
	var requestedData []string
	values := reflect.ValueOf(filter)
	types := values.Type()
	for i := 0; i < values.NumField(); i++ {
		if reflect.TypeOf(values.Field(i)).Kind() == reflect.Bool {
			if values.Field(i).Bool() {
				if types.Field(i).Name == "Group" {
					requestedData = append(requestedData, "\"group\"")
				}
				requestedData = append(requestedData, types.Field(i).Name)
			}
		}
	}
	return filter.Limit, filter.Page, requestedData
}
*/
