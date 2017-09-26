package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"mvdan.cc/xurls"
)

var codec Codec

// Config used for database configuration
type Config struct {
	DbName     string `json:"DB_NAME"`
	DbUser     string `json:"DB_USER"`
	DbPassword string `json:"DB_PASSWORD"`
}

// URL is used for database migration
type URL struct {
	gorm.Model
	Source string `json:"source"`
}

// Codec is used for encoding / decoing our source url id
type Codec interface {
	Encode(string) string
	Decode(string) (string, error)
}

// Base64 is main struct for our base64 encoding
type Base64 struct {
	e *base64.Encoding
}

// ErrorResponse should throw status and message
type ErrorResponse struct {
	Status  int
	Message string
}

// Base64Codec is used to initialize our base64
func Base64Codec() Base64 {
	return Base64{base64.URLEncoding}
}

// Encode will convert our source id into Base64
func (b Base64) Encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.Replace(str, "=", "", -1)
}

// Decode will convert Base64 to source url id so then we can find source url in database
func (b Base64) Decode(s string) (string, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	str, err := base64.URLEncoding.DecodeString(s)
	return string(str), err
}

func main() {
	db := database()
	db.AutoMigrate(&URL{})

	codec = Base64Codec()

	r := mux.NewRouter()
	r.HandleFunc("/", handleCreate).Methods("POST")
	r.HandleFunc("/{id}", handleRedirect)
	http.ListenAndServe(":3000", r)
}

func database() *gorm.DB {
	var configfile = "./db.json"
	var config Config
	file, err := os.Open(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("mysql", config.DbUser+":"+config.DbPassword+"@/"+config.DbName+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	source := r.FormValue("url")
	isURL := xurls.Relaxed().FindString(source)

	if len(source) == 0 {
		handleError(400, "Please fill URL param!", w)
		return
	}

	if len(isURL) == 0 {
		handleError(400, "URL not valid!", w)
		return
	}

	url := URL{Source: source}
	db := database()
	db.Save(&url)

	hash := fmt.Sprint(url.ID)
	mapResult := map[string]string{"url": "http://localhost:3000/" + codec.Encode(hash)}
	result, _ := json.Marshal(mapResult)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(result)
	return
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
	id, err := codec.Decode(mux.Vars(r)["id"])
	var url URL
	db := database()
	db.First(&url, id)

	if url.ID == 0 || err != nil {
		handleError(404, "Not found!", w)
		return
	}
	http.Redirect(w, r, url.Source, 301)
}

func handleError(code int, message string, w http.ResponseWriter) {
	mapResult := ErrorResponse{Status: code, Message: message}
	result, _ := json.Marshal(mapResult)
	w.WriteHeader(code)
	w.Write(result)
}
