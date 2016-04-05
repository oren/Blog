# Practical Golang: Event multicast/subscription service

## Introduction

In our microservice architectures we always need a method for communicating between services. There are various ways to achieve this. Few of them are, but are not limited to: *Remote Procedure Call*, *REST API's*, *message BUSses*. In this comprehensive tutorial we'll write a service, which you can use to distribute messages/events across your system.

### Design

How will it work? It will accept registering subscribers (other microservices). Whenever it gets a message from a microservice, it will send it further to all subscribers, using a REST call to the other microservices /event URL.

Subscribers will need to call a *keep-alive* URL regularly, otherwise they will get removed from the subscriber list. This protects us from sending messages to too many ghost subscribers.

## Implementation

Let's start with a basic structure. We'll define the **API** and set up our two main data structures:
1. The ***subscriber list*** with their register/lastKeepAlive dates.
2. The ***mutex*** controlling access to our subscriber list.

```go
package main

import (
	"net/http"
	"time"
	"sync"
	"fmt"
	"net/url"
	"io/ioutil"
	"bytes"
)

var registeredServiceStorage map[string]time.Time
var serviceStorageMutex sync.RWMutex

func main() {
	registeredServiceStorage = make(map[string]time.Time)
	serviceStorageMutex = sync.RWMutex{}

	http.HandleFunc("/registerAndKeepAlive", registerAndKeepAlive)
	http.HandleFunc("/deregister", deregister)
	http.HandleFunc("/sendMessage", handleMessage)
	http.HandleFunc("/listSubscribers", handleSubscriberListing)

	go killZombieServices()
	http.ListenAndServe(":3000", nil)
}

func registerAndKeepAlive(w http.ResponseWriter, r *http.Request) {
}

func deregister(w http.ResponseWriter, r *http.Request) {
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
}

func sendMessageToSubscriber(data []byte, address string) {
}

func handleSubscriberListing(w http.ResponseWriter, r *http.Request) {
}

func killZombieServices() {
}
```

We initialize our subscriber list and mutex, and also launch, on another thread, a function that will regularly delete *ghost subscribers*.

So far so good!
We can now start getting into each functions implementation.

We can begin with the registerAndKeepAlive which does both things. Registering a new subscriber, or updating an existing one. This works because in both cases we just update the map entry with the subscriber address to contain the current time.

```go
func registerAndKeepAlive(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//Subscriber registration
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}
```

The register function should be called with a **POST** request. That's why the first thing we do, is checking if the method is right, otherwise we answer with an error. If it's ok, then we register the client:

```go
if r.Method == http.MethodPost {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:", err)
		return
	}
	if len(values.Get("address")) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error:","Wrong input address.")
		return
	}

}
```

We check if the *URL arguments* are correct, and finally register the subscriber:

```go
if len(values.Get("address")) == 0 {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:","Wrong input address.")
	return
}

serviceStorageMutex.Lock()
registeredServiceStorage[values.Get("address")] = time.Now()
serviceStorageMutex.Unlock()

fmt.Fprint(w, "success")
```

Awesome!

Let's now implement the function which shall delete the entry when the subscriber wants to deregister.

```go
func deregister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		values, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:", err)
			return
		}
		if len(values.Get("address")) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, "Error:","Wrong input address.")
			return
		}

		//Subscriber deletion will come here

	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only DELETE accepted")
	}
}
```

Again do we check if the *request method* is good and if the *address* argument is correct. If that's the case, then we can remove this client from our subscriber list.

```go
if len(values.Get("address")) == 0 {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, "Error:","Wrong input address.")
	return
}

serviceStorageMutex.Lock()
delete(registeredServiceStorage, values.Get("address"))
serviceStorageMutex.Unlock()

fmt.Fprint(w, "success")
```

Now it's time for the main functionality. Namely handling messages and sending them to all subscribers:

```go
func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}
```

As usual, we check if the request method is correct.

Then, we read the data we got, so we can pass it to multiple concurrent sending functions.

```go
if r.Method == http.MethodPost {

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	//...
}
```

We then lock the mutex ***for read***. That's important so that we can handle huge amounts of messages efficiently. Basically, it means that we allow others to read while we are reading, because concurrent reading is supported by maps. We can use this unless there's no one modifying the map.

While we lock the map for read, we check the list of subscribers we have to send the message to, and start concurrent functions that will do the sending. As we don't want to lock the map for the entire sending time, we only need the addresses.

```go
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}

	serviceStorageMutex.RLock()
	for address, _ := range registeredServiceStorage {
		go sendMessageToSubscriber(data, address)
	}
	serviceStorageMutex.RUnlock()

	fmt.Fprint(w, "success")
```

Which means we now have to implement the *sendMessageToSubscriber(...)* function.

It's pretty simple, we just make a post, and print an error if it happened.

```go
func sendMessageToSubscriber(data []byte, address string) {
	_, err := http.Post("http://" + address + "/event", "", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
	}
}
```

It's important to notice, that we have to create a *buffer* from the data, as the *http.Post(...)* function needs a reader type data structure.

We'll also implement the function which makes it possible to list all the subscribers. Mainly for debugging purposes. There's nothing new in it. We check if the method is alright, lock the mutex for read, and finally print the map with a correct format of the register time.

```go
func handleSubscriberListing(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		serviceStorageMutex.RLock()

		for address, registerTime := range registeredServiceStorage {
			fmt.Fprintln(w, address, " : ", registerTime.Format(time.RFC3339))
		}

		serviceStorageMutex.RUnlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only GET accepted")
	}
}
```

Now there's only one function left. The one that will make sure no ghost services stay for too long. It will check all the services once per minute. This way we're making it cheap on performance:

```go
func killZombieServices() {
	t := time.Tick(1 * time.Minute)

	for range t {
	}
}
```

This is a nice way to launch the code every minute. We create a channel which will send us the time every minute, and range over it, ignoring the received values.

We can now get the check and remove working.

```go
for range t {
	timeNow := time.Now()
	serviceStorageMutex.Lock()
	for address, timeKeepAlive := range registeredServiceStorage {
		if timeNow.Sub(timeKeepAlive).Minutes() > 2 {
			delete(registeredServiceStorage, address)
		}
	}
	serviceStorageMutex.Unlock()
}
```

We just range over the subscribers and delete those that haven't kept their subscription alive.

To add to that, if you wanted you could first make a read-only pass over the subscribers, and immediately after that, make a **write-locked** deletion of the ones you found. This would allow others to read the map while you're finding subscribers to delete.

## Conclusion

That's all! Have fun with creating an infrastructure based on such a service!
