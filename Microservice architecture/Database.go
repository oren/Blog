package main

import (
	"net/http"
	"net/url"
	"fmt"
)

func main() {
	http.HandleFunc("/get", getById)
	http.HandleFunc("/newTask", newTask)
	http.HandleFunc("/getLastNotFinished", getLastNotFinished)
	http.HandleFunc("/getLastNotStarted", getLastNotStarted)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/set", setById)
	http.ListenAndServe(":3000", nil)
}

func getById(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.Write(err)
		return
	}
	if len(values.Get("id")) != 1 {
		fmt.Fprint(w, "Wrong input")
		return
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







