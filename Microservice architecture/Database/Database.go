package main

import (
	"net/http"
	"net/url"
	"fmt"
	"sync"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"time"
	"os"
)

type Task struct {
	Id int `json:"id"`
	State int `json:"state"`
}

var datastore map[int]Task
var datastoreMutex sync.Mutex
var oldestNotFinishedTask int // remember to account for potential int overflow in production. Use something bigger.
var oNFTMutex sync.Mutex

func main() {

	if !registerInKVStore() {
		return
	}

	datastore = make(map[int]Task)
	datastoreMutex = sync.Mutex{}
	oldestNotFinishedTask = 0
	oNFTMutex = sync.Mutex{}

	http.HandleFunc("/getById", getById)
	http.HandleFunc("/newTask", newTask)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/setById", setById)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":3001", nil)
}

func getById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}

		id, err := strconv.Atoi(string(values.Get("id")))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		datastoreMutex.Lock()
		bIsInError := err != nil || id >= len(datastore) // Reading the length of a slice must be done in a synchronized manner. That's why the mutex is used.
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
			fmt.Fprint(w, err)
			return
		}

		fmt.Fprint(w, string(response))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}

func newTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		datastoreMutex.Lock()
		taskToAdd := Task{
			Id: len(datastore),
			State: 0,
		}
		datastore[taskToAdd.Id] = taskToAdd
		datastoreMutex.Unlock()

		fmt.Fprint(w, taskToAdd.Id)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

		bErrored := false

		datastoreMutex.Lock()
		if len(datastore) == 0 {
			bErrored = true
		}
		datastoreMutex.Unlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: No non-started task.")
			return
		}

		taskToSend := Task{Id: -1, State: 0}

		oNFTMutex.Lock()
		datastoreMutex.Lock()
		for i := oldestNotFinishedTask; i < len(datastore); i++ {
			if datastore[i].State == 2 && i == oldestNotFinishedTask {
				oldestNotFinishedTask++
				continue
			}
			if datastore[i].State == 0 {
				datastore[i] = Task{Id: i, State: 1}
				taskToSend = datastore[i]
				break
			}
		}
		datastoreMutex.Unlock()
		oNFTMutex.Unlock()

		if taskToSend.Id == -1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: No non-started task.")
			return
		}

		myId := taskToSend.Id

		go func() {
			time.Sleep(time.Second * 120)
			datastoreMutex.Lock()
			if datastore[myId].State == 1 {
				datastore[myId] = Task{Id: myId, State: 0}
			}
			datastoreMutex.Unlock()
		}()

		response, err := json.Marshal(taskToSend)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		fmt.Fprint(w, string(response))
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func finishTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Wrong input")
			return
		}

		id, err := strconv.Atoi(string(values.Get("id")))

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		updatedTask := Task{Id: id, State: 2}

		bErrored := false

		datastoreMutex.Lock()
		if datastore[id].State == 1 {
			datastore[id] = updatedTask
		} else {
			bErrored = true
		}
		datastoreMutex.Unlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Wrong input")
			return
		}

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func setById(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		taskToSet := Task{}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		err = json.Unmarshal([]byte(data), &taskToSet)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		bErrored := false
		datastoreMutex.Lock()
		if taskToSet.Id >= len(datastore) || taskToSet.State > 2 || taskToSet.State < 0 {
			bErrored = true
		} else {
			datastore[taskToSet.Id] = taskToSet
		}
		datastoreMutex.Unlock()

		if bErrored {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: Wrong input")
			return
		}

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func list(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		datastoreMutex.Lock()
		for key, value := range datastore {
			fmt.Fprintln(w, key, ": ", "id:", value.Id, " state:", value.State)
		}
		datastoreMutex.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}

func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return false
	}
	databaseAddress := os.Args[1] // The address of itself
	keyValueStoreAddress := os.Args[2]

	response, err := http.Post(keyValueStoreAddress + "/set?key=databaseAddress&value=" + databaseAddress, "", nil)
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




