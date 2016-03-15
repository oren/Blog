package main

import (
	"net/http"
	"net/url"
	"fmt"
	"sync"
	"strconv"
	"encoding/json"
)

type Task struct {
}

var datastore []Task
var datastoreMutex sync.Mutex
var oldestNotFinishedTask int // remember to account for potential int overflow in production. Use something bigger.
var oNFTMutex sync.Mutex

func main() {

	datastore = make([]Task, 0)
	datastoreMutex = sync.Mutex{}
	oldestNotFinishedTask = 0
	oNFTMutex = sync.Mutex{}

	http.HandleFunc("/getById", getById)
	http.HandleFunc("/newTask", newTask)
	http.HandleFunc("/getLastNotFinished", getLastNotFinished)
	http.HandleFunc("/getLastNotStarted", getLastNotStarted)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/set", setById)
	http.ListenAndServe(":3001", nil)
}

func getById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.Write(err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}

		id, err := strconv.Atoi(string(values.Get("id")))
		datastoreMutex.Lock()
		bIsInError := err != nil || id > len(datastore) // Reading the length of a slice msut be done in a synchronized manner. That's why the mutex is used.
		datastoreMutex.Unlock()

		if bIsInError {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}

		datastoreMutex.Lock()
		value := datastore[id]
		datastoreMutex.Unlock()

		response, err := json.Marshal(value)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(err)
			return
		}

		fmt.Fprint(w, response)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}

func newTask(w http.ResponseWriter, r *http.Request) {
}

func getLastNotFinished(w http.ResponseWriter, r *http.Request) {
}

func getLastNotStarted(w http.ResponseWriter, r *http.Request) {
}

func finishTask(w http.ResponseWriter, r *http.Request) {
}

func setById(w http.ResponseWriter, r *http.Request) {
}







