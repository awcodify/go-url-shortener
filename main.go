package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var codec Codec

type Url struct {
	gorm.Model
	Source string `json:"source"`
}

type Codec interface {
	Encode(string) string
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

func main() {
	db := Database()
	db.AutoMigrate(&Url{})

	codec = Base64Codec()

	r := mux.NewRouter()
	r.HandleFunc("/", handleCreate).Methods("POST")
	http.ListenAndServe(":3000", r)
}

func Database() *gorm.DB {
	db, err := gorm.Open("mysql", "username:password@/dbname?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Errorf("failed to connect database")
	}
	return db
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	source := r.FormValue("url")

	if len(source) == 0 {
		mapResult := map[string]string{"message": "url not found!"}
		result, _ := json.Marshal(mapResult)
		w.WriteHeader(400)
		w.Write(result)
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
