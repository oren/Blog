package main

import (
	"net/http"
	"os"
	"fmt"
	"io/ioutil"
)

func main() {
	http.Post("localhost:3000/registerAndKeepAlive?address=" + os.Args[1], "", nil)

	http.HandleFunc("/event", printEvent)
	http.ListenAndServe(os.Args[1], nil)
}

func printEvent(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
}