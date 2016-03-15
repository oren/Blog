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
* It will allow to get the last not finished task.
* It will allow to get the last not started task.
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

First, we should create the API and later we will add the implementations of the functionality as before with the key-value store. We will also need a global slice being our data store, a variable pointing to the oldest not finished task, and mutexes for accessing the datastore and pointer.

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
	http.HandleFunc("/getLastNotFinished", getLastNotFinished)
	http.HandleFunc("/getLastNotStarted", getLastNotStarted)
	http.HandleFunc("/finishTask", finishTask)
	http.HandleFunc("/set", setById)
	http.ListenAndServe(":3001", nil)
}

func getById(w http.ResponseWriter, r *http.Request) {
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
```

We check if the ***GET*** method has been used. Later we parse the *id* argument and check if it's proper. We then get the *id* as an **int** using the *strconv.Atoi* function. Next we make sure it is not out of bounds for our *datastore*, which we have to do using *mutexes* because we're accessing a slice which could be accessed from another thread. If everything is ok, then, again using *mutexes*, we get the task using the *id*.

After that we use the *JSON* library to marshal our struct into a *JSON object* and if that finishes without problems we send the *JSON object* to the client.
