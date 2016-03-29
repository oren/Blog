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

		serviceStorageMutex.Lock()
		registeredServiceStorage[values.Get("address")] = time.Now()
		serviceStorageMutex.Unlock()

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

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

		serviceStorageMutex.Lock()
		delete(registeredServiceStorage, values.Get("address"))
		serviceStorageMutex.Unlock()

		fmt.Fprint(w, "success")
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only DELETE accepted")
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {

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
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Only POST accepted")
	}
}

func sendMessageToSubscriber(data []byte, address string) {
	_, err := http.Post("http://" + address + "/event", "", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
	}
}

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

func killZombieServices() {
}
