# go-url-shortener

Just a simple url shortener using Base64.
This simple application have ability to:
* check is param filled or not?
* check is url valid or not?
* return 404 when url not found in database

# How to use
* edit db.conf (we are using MySQL)
* just run `go run main.go`

# Routing

| Endpoint    | Method | Response        | Feature                          |
| ----------- |:------:|-----------------| ---------------------------------|
| /?url={url} | POST   | JSON            | url validation, mandatory param  |
| /{id}       | GET    | 301 Redirection |                                  |
