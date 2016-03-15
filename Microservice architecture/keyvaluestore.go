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
	keyValueStore = make(map[string]string)
	kVStoreMutex = sync.Mutex{}
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":3000", nil)
}

func get(w http.ResponseWriter, r *http.Request) {
	if(r.Method == http.MethodGet) {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("key")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input key.")
			return
		}

		kVStoreMutex.Lock()
		value := keyValueStore[string(values.Get("key"))]
		kVStoreMutex.Unlock()

		fmt.Fprint(w, value)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}

func set(w http.ResponseWriter, r *http.Request) {
	if(r.Method == http.MethodPost) {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("key")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "Wrong input key.")
			return
		}
		if len(values.Get("value")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", "Wrong input value.")
			return
		}

		kVStoreMutex.Lock()
		keyValueStore[string(values.Get("key"))] = string(values.Get("value"))
		kVStoreMutex.Unlock()

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted.")
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	if(r.Method == http.MethodGet) {
		kVStoreMutex.Lock()
		for key, value := range keyValueStore {
			fmt.Fprintln(w, key, ":", value)
		}
		kVStoreMutex.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}


