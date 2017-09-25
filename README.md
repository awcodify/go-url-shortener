# go-url-shortener

Just a simple url shortener using Base64.

# How to use
* edit db.conf (we are using MySQL)
* just run `go run main.go`

# Routing

| Endpoint    | Method | Response        |
| ----------- |:------:|-----------------|
| /?url={url} | POST   | JSON            |
| /{id}       | GET    | 301 Redirection |
