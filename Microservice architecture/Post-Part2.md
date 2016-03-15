# Web app using Microservices in Go: Part 2 - Implementation

## Introduction

In this part we will implement part of the microservices needed for our web app. We will implement the:
* key-value store
* Database

## The key-value store

### Design

The design hasn't changed much. We will save the key-value pairs as a global map, and create a global mutex for concurrent access. We'll also add the ability to list all key-value pairs for debugging/analytical purposes.

First, let's create the structure:

```go
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
}

func set(w http.ResponseWriter, r *http.Request) {
}

func list(w http.ResponseWriter, r *http.Request) {
}

```

And now let's dive into the implementation.

First, we should add parameter parsing in the get function and verify that the key parameter is right.

```go
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
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted.")
	}
}
```

The *key* shouldn't have a length of 0, hence the length check. We also check if the method is GET, if it isn't we print it and set the status code to ***bad request***.
We answer with an explicit ***Error:*** before each error message so it doesn't get misinterpreted by the client as a value.

Now, let's access our map and send back a response:

```go
if len(values.Get("key")) == 0 {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:","Wrong input key.")
	return
}

kVStoreMutex.Lock()
value := keyValueStore[string(values.Get("key"))]
kVStoreMutex.Unlock()

fmt.Fprint(w, value)
```

We copy the value into a variable so that we don't block the map while sending back the response.

Now let's create the set function, it's actually pretty similar.

```go
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
```

The only difference is that we also check if there is a right value parameter and check if the method is POST.

Now we can add the implementation of the list function which is also pretty simple:

```go
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
```

It just ranges over the map and prints everything. Simple yet effective.

## The database

### Design

After thinking through the design, I decided that it would be better if the database generated the task *Id*'s. This will also make it easier to get the last non-finished task and generate consecutive *Id*'s

How it will work:
* It will save new tasks assigning consecutive *Id*'s.
* It will remember the oldest not finished task.
* It will allow to get a new task to do.
* It will allow to get a task by *Id*.
* It will allow to set a task by *Id*.
* The state will be represented by an int:
  * 0 - not started
  * 1 - in progress
  * 2 - finished
* It will change the state of a task to *not started* if it's been too long *in progress*. (maybe someone started to work on it but has crashed)
* It will allow to list all tasks for debugging/analytical purposes.

![Database microservice post](https://www.lucidchart.com/publicSegments/view/4cf0690e-3dbb-42d9-befd-4a6efaaf6f72/image.png)

### Implementation

First, we should create the API and later we will add the implementations of the functionality as before with the key-value store. We will also need a global map being our data store, a variable pointing to the oldest not finished task, and mutexes for accessing the datastore and pointer.

```go
package main

import (
	"net/http"
	"net/url"
	"fmt"
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
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/setById", setById)
	http.HandleFunc("/list", list)
	http.ListenAndServe(":3001", nil)
}

func getById(w http.ResponseWriter, r *http.Request) {
}

func newTask(w http.ResponseWriter, r *http.Request) {
}

func getNewTask(w http.ResponseWriter, r *http.Request) {
}

func finishTask(w http.ResponseWriter, r *http.Request) {
}

func setById(w http.ResponseWriter, r *http.Request) {
}

func list(w http.ResponseWriter, r *http.Request) {

}
```

We also already declared the ***Task*** type which we will use for storage.

//TODO

//Write the service discovery code

//to register in the key-value store

//self ip through argument

//

//

//

So far so good. Now let's implement all those functions!

First, let's implement the getById function.

```go
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
```

We check if the ***GET*** method has been used. Later we parse the *id* argument and check if it's proper. We then get the *id* as an **int** using the *strconv.Atoi* function. Next we make sure it is not out of bounds for our *datastore*, which we have to do using *mutexes* because we're accessing a map which could be accessed from another thread. If everything is ok, then, again using *mutexes*, we get the task using the *id*.

After that we use the *JSON* library to marshal our struct into a *JSON object* and if that finishes without problems we send the *JSON object* to the client.

It's also time to implement our *Task* struct:

```go
type Task struct {
	Id int `json:"id"`
	State int `json:"state"`
}
```

It's all that's needed. We also added the information the *JSON* marshaller needs.

We can now go on with implementing the *newTask* function:

```go
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
```

It's pretty small actually. Creating a new *Task* with the next id and adding it to the *datastore*. After that it sends back the new *Tasks* Id.

That means we can go on to implementing the function used to list all *Tasks*, as this helps with debugging during writing.

It's basically the same as with the key-value store:

```go
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
```

Ok, so now we will implement the function which can set the *Task* by *id*:

```go
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
```

Nothing new. We get the request and try to unmarshal it. If it succeeds we put it into the map, checking if it isn't out of bounds or if the state is invalid. If it is then we print an error, otherwise we print *success*.

If we already have this we can now implement the finish task function, because it's very simple:

```go
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
```

It's pretty similar to the *getById* function. The difference here is that here we update the state and only if it is currently in progress.

And now to one of the most interesting functions. The *getNewTask* function. It has to handle updating the oldest known finished task, and it also needs to handle the situation when someone takes a task but crashes during work. This would lead to a ghost task forever being *in progress*. That's why we'll add functionality which after 120 seconds from starting a task will set it back to *not finished*:

```go
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
			fmt.Fprint(w, "Error: No non-finished task.")
			return
		}

		taskToSend := Task{Id: -1, State: 0}

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

		if taskToSend.Id == -1 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error: No non-finished task.")
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
```

First we try to find the oldest task that hasn't started yet. By the way we update the oldestNotFinishedTask variable. If a task is finished and is pointed on by the variable, the variable get's incremented. If we find something that's not started, then we break out of the loop and send it back to the user setting it to *in progress*. However, on the way we start a function on another thread that will change the state of the task back to *not started* if it's in progress for more than 120 seconds.
