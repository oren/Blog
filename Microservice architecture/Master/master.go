package main

import (
	"os"
	"fmt"
	"net/http"
	"io/ioutil"
)

func main() {
	if !registerInKVStore() {
		return
	}

	http.HandleFunc("/new", newImage)
	http.HandleFunc("/get", getImage)
	http.HandleFunc("/isReady", isReady)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/registerTaskFinished", registerTaskFinished)
}

func newImage(w http.ResponseWriter, r *http.Request) {
}

func getImage(w http.ResponseWriter, r *http.Request) {
}

func isReady(w http.ResponseWriter, r *http.Request) {
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
}

func registerTaskFinished(w http.ResponseWriter, r *http.Request) {
}

func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return false
	}
	masterAddress := os.Args[1] // The address of itself
	keyValueStoreAddress := os.Args[2]

	response, err := http.Post("http://" + keyValueStoreAddress + "/set?key=masterAddress&value=" + masterAddress, "", nil)
	if err != nil {
		fmt.Println(err)
		return false
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: Failure when contacting key-value store: ", string(data))
		return false
	}
	return true
}