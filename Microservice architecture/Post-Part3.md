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

		_, err = strconv.Atoi(values.Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input id.")
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

We create a file in the tmp/*state* directory with the right *id*.  Another thing we do is check if the *id* really is a valid int. We parse it to an int, to see if it succeeds and if it does then we use it, as a string.

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

		_, err = strconv.Atoi(values.Get("id"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input id.")
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

type Task struct {
	Id int `json:"id"`
	State int `json:"state"`
}

var databaseLocation string
var storageLocation string

func main() {
	if !registerInKVStore() {
		return
	}

	http.HandleFunc("/new", newImage)
	http.HandleFunc("/get", getImage)
	http.HandleFunc("/isReady", isReady)
	http.HandleFunc("/getNewTask", getNewTask)
	http.HandleFunc("/registerTaskFinished", registerTaskFinished)
	http.ListenAndServe(":3003", nil)
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

It's the structure of the **API** and the mechanics to register in the *k/v store*.

We also need to get the storage and database locations in the main function:

```go
if !registerInKVStore() {
		return
	}
	keyValueStoreAddress = os.Args[2]

	response, err := http.Get("http://" + keyValueStoreAddress + "/get?key=databaseAddress")
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: can't get database address.")
		fmt.Println(response.Body)
		return
	}
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	databaseLocation = string(data)

	response, err = http.Get("http://" + keyValueStoreAddress + "/get?key=storageAddress")
	if response.StatusCode != http.StatusOK {
		fmt.Println("Error: can't get storage address.")
		fmt.Println(response.Body)
		return
	}
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	storageLocation = string(data)
```

Now we can start implementing all the functionality!

Let's start with the *newImage* function as it contains a good bit of code and mechanics which will be again used in the other funtions.
Here's the beginning:

```go
func newImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		response, err := http.Post("http://" + databaseLocation + "/newTask", "text/plain", nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		id, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}
```

As usual we check if the method is right. Next we register a new *Task* in the database and get and *Id*.

We now use this to send the image to the storage:

```go
id, err := ioutil.ReadAll(response.Body)
if err != nil {
	fmt.Println(err)
	return
}

_, err = http.Post("http://" + storageLocation + "/sendImage?id=" + string(id) + "&state=working", "image", r.Body)
if err != nil {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:", err)
	return
}
fmt.Fprint(w, string(id))
```

That's it. The new task will be created, the storage will get a file into the working directory with the name of the file being the *id*, and the client gets back the *id*. The important thing here is that we need the raw image in the request. The user form has to be parsed in the frontend service.

Now we can create the function which just checks if a *Task* is ready:

```go
func isReady(w http.ResponseWriter, r *http.Request) {
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
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}
```

We first have to verify all the parameters and the request method. Next we can ask the database for the *Task* requested:

```go
if len(values.Get("id")) == 0 {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Wrong input")
	return
}

response, err := http.Get("http://" + databaseLocation + "/getById?id=" + values.Get("id"))
if err != nil {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:", err)
	return
}
data, err := ioutil.ReadAll(response.Body)
if err != nil {
	fmt.Println(err)
	return
}
```

We also read the response immediately. Now we can parse the *Task* and respond to the client:

```go
if err != nil {
	fmt.Println(err)
	return
}

myTask := Task{}
json.Unmarshal(data, &myTask)

if(myTask.State == 2) {
	fmt.Fprint(w, "1")
} else {
	fmt.Fprint(w, "0")
}
```

So now we can implement the last client facing interface, the *getImage* function:

```go
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
} else {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error: Only GET accepted")
}
```

Here we verified the request and now we need to get the image from the storage system, and just copy the response to our client:

```go
if len(values.Get("id")) == 0 {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Wrong input")
	return
}

response, err := http.Get("http://" + storageLocation + "/getImage?id=" + values.Get("id") + "&state=finished")
if err != nil {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:", err)
	return
}

_, err = io.Copy(w, response.Body)
if err != nil {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:", err)
	return
}
```

That's it! The client facing interface is finished!

### Implementing the worker facing interface

Now we have to implement the functions to serve the workers.

Both functions will basically be just direct routes to the database and back, so now let's write 'em too:

```go
func getNewTask(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		response, err := http.Post("http://" + databaseLocation + "/getNewTask", "text/plain", nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		_, err = io.Copy(w, response.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func registerTaskFinished(w http.ResponseWriter, r *http.Request) {
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

		response, err := http.Post("http://" + databaseLocation + "/finishTask?id=" + values.Get("id"), "test/plain", nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}

		_, err = io.Copy(w, response.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}
```

There's not much to explain. They are both just passing further the request and responding with what they get.

You could think the workers should communicate directly with the database to get new *Tasks*. And with the current implementation it would work perfectly. However, if we wanted to add some functionality the *master* wanted to do for each of those requests it would be hard to implement. So this way is very extensible, and that's nearly always what we want.

## Conclusion

Now we have finished the *Master* and the *Storage system*. We now have the dependencies to create the workers and frontend which we will implement in the next part. As always I encourage you to comment about your opinion. Have fun extending the system to do what you want to achieve!

[1]:https://jacobmartins.com/2016/03/16/web-app-using-microservices-in-go-part-2-kv-store-and-database/
