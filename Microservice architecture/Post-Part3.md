# Web app using Microservices in Go: Part 3 - Storage and Master

## Introduction

In this part we will implement the next part of the microservices needed for our web app. We will implement the:
* Storage system
* Master

This way we will have the *Master API* ready when we'll be writing the slaves/workers and the frontend. And we'll already have the database, k/v store and storage when writing the master. SO every time we write something we'll already have all its dependencies.

## The storage system

Ok, this one will be pretty easy to write. Just handling files. Let's build the basic structure, which will include a function to register in our k/v store. For reference how it works check out the [previous part][1]. So here's the basic structure:

```go
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"net/url"
	"io"
)

func main() {
	if !registerInKVStore() {
		return
	}
	http.HandleFunc("/sendImage", receiveImage)
	http.HandleFunc("/getImage", serveImage)
	http.ListenAndServe(":3002", nil)
}

func receiveImage(w http.ResponseWriter, r *http.Request) {
}

func serveImage(w http.ResponseWriter, r *http.Request) {
}

func registerInKVStore() bool {
	if len(os.Args) < 3 {
		fmt.Println("Error: Too few arguments.")
		return false
	}
	storageAddress := os.Args[1] // The address of itself
	keyValueStoreAddress := os.Args[2]

	response, err := http.Post("http://" + keyValueStoreAddress + "/set?key=storageAddress&value=" + storageAddress, "", nil)
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
```

So now we'll have to handle the file serving/uploading. We will use a state *url argument* to specify if we are using the not yet finished (aka *working*) directory, or the *finished* one.

So first let's write the ***receiveImage*** function which is there to get the files from clients:

```go
func receiveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input id.")
			return
		}
		if values.Get("state") != "working" && values.Get("state") != "finished" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input state.")
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}
```

Here we check if the request method is **POST**, if there is an *id*, and if the state is *working* or *finished*.

Next we can create the file and put in the image:

```go
if values.Get("state") != "working" && values.Get("state") != "finished" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input state.")
			return
		}

		file, err := os.Create("/tmp/" + values.Get("state") + "/" + values.Get("id") + ".png")
		defer file.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		_, err = io.Copy(file, r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		fmt.Fprint(w, "success")
```

We create a file in the tmp/*state* directory with the right *id*.

we use the *io.Copy* function to put all the data from the *request* to the file. That means that the *body* of our request should be a raw image.

Next we can write the function to serve images which is pretty similar:

```go
func serveImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("id")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input id.")
			return
		}
		if values.Get("state") != "working" && values.Get("state") != "finished" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input state.")
			return
		}

		file, err := os.Open("/tmp/" + values.Get("state") + "/" + values.Get("id") + ".png")
		defer file.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		_, err = io.Copy(w, file)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}
```

Instead of creating the file, we open it. Instead of copying to the file we copy from it. And we check if the method is **GET**.

That's it. We've got a storage service which saves and servers raw image files. Now we can get to the master!

## The master

We now have all the dependencies the master needs. So let's write it now. Here's the basic structure:

```go
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
```

It's the structure of the **API** and the me


[1]:https://jacobmartins.com/2016/03/16/web-app-using-microservices-in-go-part-2-kv-store-and-database/
