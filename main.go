package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Song struct {
	Id          int    `gorm:"type:serial;primaryKey"`
	Group       string `json:"group" gorm:"type:text"`
	Name        string `json:"name" gorm:"type:text"`
	ReleaseDate string `json:"releasedate" gorm:"column:releasedate;type:text"`
	Text        string `json:"text" gorm:"type:text"`
	Link        string `json:"link" gorm:"type:text"`
}

type DataFilter struct {
	Group       bool `json:"group"`
	Name        bool `json:"name"`
	ReleaseDate bool `json:"releasedate"`
	Text        bool `json:"text"`
	Link        bool `json:"link"`
}

var dblog string
var inforeq string
var db *gorm.DB

// @title EffectiveMobileAPI
// @version 1.0
// @host localhost:8080
// @schemes http
// @BasePath /
func init() {
	if err := godotenv.Load("go.env"); err != nil {
		panic(err)
	}
	var exist bool
	dblog, exist = os.LookupEnv("dblog")
	log.Printf("DB connection parameters: %v, Exist: %v", dblog, exist)
	inforeq, exist = os.LookupEnv("inforeq")
	log.Printf("Fortified connection link: %v, Exist: %v", inforeq, exist)
}
func main() {
	db = Migrate()
	mux := http.NewServeMux()
	mux.HandleFunc("/getlibdata", GetLibData)
	mux.HandleFunc("/getsongtext", GetSongText)
	mux.HandleFunc("/deletesong", DeleteSong)
	mux.HandleFunc("/changesong", ChangeSong)
	mux.HandleFunc("/addsong", AddSong)
	handler := cors.Default().Handler(mux)
	http.ListenAndServe("localhost:8080", handler)
}

// @Summary Получить список песен
// @Description Возвращает список всех песен
// @Param Filter header string true "DataFilter"
// @Param limit query int true "Songs per page"
// @Param page query int true "Page"
// @Success 200 {array} interface{}
// @Router /getlibdata [get]
func GetLibData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(404)
		return
	}
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var datafilter DataFilter
	for _, value := range strings.Split(r.Header["Filter"][0], ", ") {
		switch value {
		case "group":
			datafilter.Group = true
		case "name":
			datafilter.Name = true
		case "releasedate":
			datafilter.ReleaseDate = true
		case "text":
			datafilter.Text = true
		case "link":
			datafilter.Link = true
		}
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	log.Printf("Songs per page: %v, Current page: %v\n", limit, page)
	res, err := db.Query(fmt.Sprintf("select * from songs order by id limit %v offset %v", limit, (page-1)*limit))
	if err != nil {
		panic(err)
	}
	var songs []Song
	for res.Next() {
		var song Song
		if err = res.Scan(&song.Id, &song.Group, &song.Name, &song.ReleaseDate, &song.Text, &song.Link); err != nil {
			panic(err)
		}
		songs = append(songs, song)
	}
	log.Printf("Songs to show: %v", songs)
	body := []interface{}{}
	for i := range songs {
		song := make(map[string]interface{})
		songValues := reflect.ValueOf(songs[i])
		songTypes := songValues.Type()
		filterValues := reflect.ValueOf(datafilter)
		filterTypes := filterValues.Type()
		for i := 0; i < songValues.NumField(); i++ {
			for j := 0; j < filterValues.NumField(); j++ {
				if filterTypes.Field(j).Name == songTypes.Field(i).Name {
					if filterValues.Field(j).Bool() {
						song[songTypes.Field(i).Name] = songValues.Field(i).String()
					}
				}
			}
		}
		body = append(body, song)
	}
	bodyjson, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	w.Write([]byte(bodyjson))
}

// @Summary Получение песни с пагинацией
// @Produce json
// @Param name query string true "Song name"
// @Param group query string true "Song group"
// @Param limit query int true "Songs per page"
// @Param page query int true "Page"
// @Success 200 {object} string
// @Router /getsongtext [get]
func GetSongText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(404)
		return
	}
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var song Song
	song.Group = r.URL.Query().Get("group")
	song.Name = r.URL.Query().Get("name")
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	log.Printf("Song to find: %v", song)
	log.Printf("Couplets to show: %v, Current page: %v", limit, page)
	row := db.QueryRow(fmt.Sprintf("select text from songs where \"group\"='%v' and name='%v'", song.Group, song.Name))
	var songText string
	err = row.Scan(&songText)
	log.Printf("Result text: %v", songText)
	if err == sql.ErrNoRows {
		fmt.Fprintln(w, "No songs with such group and name")
		return
	}
	if err != nil {
		panic(err)
	}
	songTextPagination := strings.Split(songText, `\n\n`)
	var body string
	w.WriteHeader(200)
	for i := (page - 1) * limit; i < page*limit; i++ {
		songTextPagination[i] += "\n"
		body += strings.Replace(songTextPagination[i], `\n`, "\n", -1)
	}
	bodyjson, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	w.WriteHeader(200)
	w.Write([]byte(bodyjson))
}

// @Summary Удаление песни
// @Accept json
// @Produce json
// @Param song body Song true "Song name and group"
// @Success 200
// @Router /deletesong [delete]
func DeleteSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		return
	}
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var song Song
	err = json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		w.WriteHeader(400)
		panic(err)
	}
	log.Printf("Song to delete: %v", song)

	res, err := db.Exec(fmt.Sprintf("delete from songs where \"group\"='%v' and name='%v'", song.Group, song.Name))
	if err != nil {
		w.WriteHeader(400)
		log.Panic(err)
	}
	log.Printf("Result: %v", res)
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Panic(err)
	}
	w.WriteHeader(200)
	if rowsAffected == 0 {
		fmt.Fprintln(w, "No songs were found")
	} else {
		fmt.Fprintln(w, "Song was deleted succesfully")
	}
}

// @Summary Изменение песни
// @Accept json
// @Produce json
// @Param song body Song true "Song name and group"
// @Success 200
// @Router /changesong [patch]
func ChangeSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		return
	}
	var song Song
	err := json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		panic(err)
	}
	song.Id = 0
	log.Printf("Song to change: %v", song)
	res := db.Model(&Song{}).Where("name = ? and \"group\" = ?", song.Name, song.Group).Updates(song)
	log.Printf("Result: %v", res.Statement)
	if res.Error != nil {
		panic(res.Error)
	}
	w.WriteHeader(200)
	if res.RowsAffected == 0 {
		fmt.Fprintln(w, "No songs to change with such group and name")
	} else {
		fmt.Fprintln(w, "Song was changed successfully")
	}
}

// @Summary Добавление песни
// @Accept json
// @Produce json
// @Param song body Song true "Song name and group"
// @Success 200
// @Router /addsong [post]
func AddSong(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}
	db, err := sql.Open("postgres", dblog)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	var song Song
	err = json.NewDecoder(r.Body).Decode(&song)
	if err != nil {
		panic(err)
	}
	log.Println("Song to add: " + song.Group + ": " + song.Name)
	songFiller := Song{}
	resp, err := http.Get(inforeq)
	if err != nil {
		log.Println("Can't access filler server, filling with base values")
		songFiller.ReleaseDate = "27.07.1987"
		songFiller.Text = "Ooh baby, dont you know I suffer?" + `\n` + "Ooh baby, can you hear me moan?" + `\n` + "You caught me under false pretenses" + `\n` + "How long before you let me go?" + `\n` + "" + `\n` + "Ooh" + `\n` + "You set my soul alight" + `\n` + "Ooh" + `\n` + "You set my soul alight"
		songFiller.Link = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	} else {
		defer resp.Body.Close()
		err = json.NewDecoder(resp.Body).Decode(&songFiller)
		if err != nil {
			panic(err)
		}
	}
	song.ReleaseDate = songFiller.ReleaseDate
	song.Text = songFiller.Text
	song.Link = songFiller.Link
	fmt.Printf("Fortified song: %v", song)
	res, err := db.Exec(fmt.Sprintf("insert into songs(\"group\", name, releasedate, text, link) values('%v', '%v', '%v', '%v', '%v')",
		song.Group, song.Name, song.ReleaseDate, song.Text, song.Link))
	if err != nil {
		panic(err)
	}
	log.Println(res)
	w.WriteHeader(200)
	fmt.Fprintln(w, "Song added successfully")
}

func Migrate() *gorm.DB {
	db, err := gorm.Open(postgres.Open(dblog), &gorm.Config{})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	err = db.AutoMigrate(&Song{})
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	fillerSong := Song{Name: "song1", Group: "group1", ReleaseDate: "rdate1", Text: "Ooh baby, don't you know I suffer?" + `\n` + "Ooh baby, can you hear me moan?" + `\n` + "You caught me under false pretenses" + `\n` + "How long before you let me go?" + `\n` + "" + `\n` + "Ooh" + `\n` + "You set my soul alight" + `\n` + "Ooh" + `\n` + "You set my soul alight", Link: "link1"}
	if result := db.Create(&fillerSong); result.Error != nil {
		panic(result.Error)
	}
	log.Println("Migration completed")
	return db
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
