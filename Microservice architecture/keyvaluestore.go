package main

import "net/http"

func main() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
}

func get(w http.ResponseWriter, r *http.Request) {
}

func set(w http.ResponseWriter, r *http.Request) {
}


