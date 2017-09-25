package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"mvdan.cc/xurls"
)

var codec Codec

type Config struct {
	DB_NAME     string
	DB_USER     string
	DB_PASSWORD string
}

type Url struct {
	gorm.Model
	Source string `json:"source"`
}

type Codec interface {
	Encode(string) string
	Decode(string) (string, error)
}

type Base64 struct {
	e *base64.Encoding
}

type ApiResponse struct {
	Status  int
	Message string
}

func Base64Codec() Base64 {
	return Base64{base64.URLEncoding}
}

func (b Base64) Encode(s string) string {
	str := base64.URLEncoding.EncodeToString([]byte(s))
	return strings.Replace(str, "=", "", -1)
}

func (b Base64) Decode(s string) (string, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	str, err := base64.URLEncoding.DecodeString(s)
	return string(str), err
}

func main() {
	db := Database()
	db.AutoMigrate(&Url{})

	codec = Base64Codec()

	r := mux.NewRouter()
	r.HandleFunc("/", handleCreate).Methods("POST")
	r.HandleFunc("/{id}", handleRedirect)
	http.ListenAndServe(":3000", r)
}

func Database() *gorm.DB {
	var configfile = "./db.conf"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open("mysql", config.DB_USER+":"+config.DB_PASSWORD+"@/"+config.DB_NAME+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	source := r.FormValue("url")
	isUrl := xurls.Relaxed().FindString(source)

	if len(source) == 0 {
		handleError(400, "Please fill URL param!", w)
		return
	}

	if len(isUrl) == 0 {
		handleError(400, "URL not valid!", w)
		return
	}

	url := Url{Source: source}
	db := Database()
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
	if err != nil {
		handleError(404, "Not found!", w)
		return
	}
	var url Url
	db := Database()
	db.First(&url, id)
	http.Redirect(w, r, url.Source, 301)
}

func handleError(code int, message string, w http.ResponseWriter) {
	mapResult := ApiResponse{Status: code, Message: message}
	result, _ := json.Marshal(mapResult)
	w.WriteHeader(code)
	w.Write(result)
}
