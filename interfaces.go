package main

import "net/http"

type DB interface {
	Add(key, value string)
	Del(key string)
	GetContent() string
	Load() error
	Save() error
	HttpAddHandler(w http.ResponseWriter, r *http.Request)
	HttpDelHandler(w http.ResponseWriter, r *http.Request)
}
