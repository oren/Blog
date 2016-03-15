# Web app using Microservices in Go: Part 2 - Implementation

## Introduction

In this part we will implement part of the microservices needed for our web app. We will implement the:
* key-value store
* Database

## The key-value store

### Design

The design hasn't changed much. We will save the key-value pairs as a global map, and create a global mutex for concurrent access. We'll also add the ability to list all key-value pairs for debugging/analytical purpouses.

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
	http.HandleFunc("/get", get)
	http.HandleFunc("/set", set)
	http.HandleFunc("/list", list)
}

func get(w http.ResponseWriter, r *http.Request) {
}

func set(w http.ResponseWriter, r *http.Request) {
}

func list(w http.ResponseWriter, r *http.Request) {
}

```

And now let's dive into the implementation.

First, we should add parameter parsing in the get function, and verify that there is only one parameter.

```go
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
```

There should be only 1 *Id*, hence the length check. We also check if the method is GET.
We answer with explicit error before each error message so it doesn't get misinterpreted by the client.

Now, let's access our map and send back a response:

```go
  if len(values.Get("key")) != 1 {
    fmt.Fprint(w, "Error")
    return
  }

  kVStoreMutex.Lock()
  value := keyValueStore[values.Get("key")[0]]
  kVStoreMutex.Unlock()

  fmt.Fprint(w, value)
```

We copy the value into a variable so that we don't block the map while sending back the response.

Now let's create the setting function, it's actually pretty similar.

```go
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
```

The only difference is that we also check if there is exactly one value parameter and check if the method is POST.

Now we can add the implementation of the list function which is also pretty simple:



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
* Change the state to *not started* if it's been too long *in progress*. (maybe someone started to work but crashed)

![Database microservice post](https://www.lucidchart.com/publicSegments/view/4cf0690e-3dbb-42d9-befd-4a6efaaf6f72/image.png)

### Implementation

First, we should create the API and later we will add the implementations.
