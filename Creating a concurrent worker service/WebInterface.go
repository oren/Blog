package ImageService

import (
	"fmt"
	"github.com/satori/go.uuid"
	"io"
	"net/http"
	"net/url"
	"os"
)

func setupWebInterface(workToDo chan string, finishedWorkMap *map[string]bool) {
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		parsedUrl, err := url.Parse(r.URL.String())
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		parsedQuery, err := url.ParseQuery(parsedUrl.RawQuery)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		globalFinishedWorkMapMutex.Lock()
		isFinished := (*finishedWorkMap)[parsedQuery["id"][0]]
		globalFinishedWorkMapMutex.Unlock()
		if isFinished {
			file, err := os.Open("/tmp/" + parsedQuery["id"][0] + ".png")
			if err == nil {
				w.Header().Set("Content-Type", "image/png")
				io.Copy(w, file)
			} else {
				fmt.Fprintln(w, err)
			}

		} else {
			fmt.Fprintln(w, "In progress or not found.")
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			newWorkId := uuid.NewV4().String()
			file, err := os.Create("/tmp/" + newWorkId + ".png")
			defer file.Close()
			if err != nil {
				fmt.Fprintln(w, err)
			}
			io.Copy(file, r.Body)
			file.Close()
			setupNewWorkJob(workToDo, newWorkId, finishedWorkMap)
			fmt.Fprintln(w, "Thank you for your submission. Job number:"+newWorkId)
		} else {
			fmt.Fprintln(w, "ERROR: Only POST accepted.")
		}
	})
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Println(err)
	}
}

func setupNewWorkJob(workToDo chan string, newWorkId string, finishedWorkMap *map[string]bool) {
	globalFinishedWorkMapMutex.Lock()
	(*finishedWorkMap)[newWorkId] = false
	globalFinishedWorkMapMutex.Unlock()
	fmt.Println("Starting dispatch...")
	workToDo <- newWorkId
	fmt.Println("Dispatched new work! :)")
}
