package main

import (
	"net/http"
	"sync"
	"net/url"
	"fmt"
)

var keyValueStore map[string]string
var kVStoreMutex sync.Mutex

func main() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/list", list)
}

func get(w http.ResponseWriter, r *http.Request) {
	if(r.Method != "GET") {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("key")) != 1 {
			fmt.Fprint(w, "Error")
			return
		}

		kVStoreMutex.Lock()
		value := keyValueStore[values.Get("key")[0]]
		kVStoreMutex.Unlock()

		fmt.Fprint(w, value)
	} else {
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}

func set(w http.ResponseWriter, r *http.Request) {
	if(r.Method != "POST") {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("key")) != 1 {
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("value")) != 1 {
			fmt.Fprint(w, "Error:", err)
			return
		}

		kVStoreMutex.Lock()
		keyValueStore[values.Get("id")[0]] = values.Get("value")[0]
		kVStoreMutex.Unlock()

		fmt.Fprint(w, "success")
	} else {
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

func list(w http.ResponseWriter, r *http.Request) {
}


