# Practical Golang: Using websockets
## Introduction

This is the first post in the *practical Golang* series. Posts in it are meant to provide short and informative introductions to various topics.

This one is a about ***websockets***, which are an awesome and easy way to provide communication between your web app and server.

Here we will use the gorilla websocket library, but you could also use a few others.

We will create two basic apps which should cover most day to day usage:
  1. A client subscribing to a server to get continues information.
  2. A client-ping server-pong app.

## Dependencies

Make sure to go get:
```
go get github.com/gorilla/websocket
```

## What's the theory?

1. The client connects to your server using his web browser.
2. He gets back the website.
3. He connects to your server through his websocket client using javascript.
4. Your server accepts it using a standard http handler.
5. You create a websocket connection from the http connection.
6. You communicate with the client.
7. The connection gets closed by one of the sides.

## Creating the subscription app

### Preparations

Next to our main go file we will need a ***html*** folder to place our html file in.

Let's name it creatively, like *index.html*

Now the contents:
```html
<!DOCTYPE HTML>
<html>
<head>

    <script type="text/javascript">
         function myWebsocketStart()
         {
               var ws = new WebSocket("ws://localhost:3000/websocket");

               ws.onmessage = function (evt)
               {
                  var myTextArea = document.getElementById("textarea1");
                  myTextArea.value = myTextArea.value + "\n" + evt.data
               };

               ws.onclose = function()
               {
                  var myTextArea = document.getElementById("textarea1");
                  myTextArea.value = myTextArea.value + "\n" + "Connection closed";
               };

         }

    </script>

</head>
<body>
<button onclick="javascript:myWebsocketStart()">Subscribe</button>
<textarea id="textarea1">MyTextArea</textarea>
</body>
</html>
```
I'll just go over it quickly as the main subject here is the go code.

We create a button and a textarea, after the user clicks the button he connects to our websocket. Whenever he receives a message, or the connection gets closed, we append the info to our textarea.

We will also need our, so creatively named, main.go file, with the basic structure and file server written:

```go
package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"fmt"
	"io/ioutil"
	"time"
	"encoding/json"
)

func main() {
	indexFile, err := os.Open("html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})
	http.ListenAndServe(":3000", nil)
}
```

Awesome, now let's create the websocket part.

### Writing the websocket code

#### Little bit of planning

Our server will create a Person object containing a name and age in seconds. Every two seconds it will send the client the current state of the person.

#### The meat

First we'll need to define our Person type:

```go
type Person struct {
	Name string
	Age  int
}
```

We'll also need to create an upgraded variable, in which we define our read and write buffer sizes.

```go
var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}
```

Now, how do we create the websocket connection? Pretty easily in fact:

```go
http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Println("Client subscribed")
}
```

That's all we need to have a client. Now let's create Bill, our person, right after we get the client subscribed:
```go
fmt.Println("Client subscribed")

myPerson := Person{
  Name:	"Bill",
  Age:	0,
}
```

Now we need the main websocket handling code, which we will wrap into an endless for loop, which we get out of only if the channel closes or Bill gets 40 seconds old.

```go
for {
  time.Sleep(2 * time.Second)
  if myPerson.Age < 40 {
    myJson, err := json.Marshal(myPerson)
    if err != nil {
      fmt.Println(err)
      return
    }
    err = conn.WriteMessage(websocket.TextMessage, myJson)
    if err != nil {
      fmt.Println(err)
      break
    }
    myPerson.Age += 2
  } else {
    conn.Close()
    break
  }
}
fmt.Println("Client unsubscribed")
```

We send the messages using conn.WriteMessage in which we specify the message type, can be binary or text, and the content. If Bill is 40 years od or more, we break out of the loop. So far so good, but what if we want bidirectional communication?

## Creating the ping-pong app

### Preparations

As before, we will need a ***html*** folder for our html file with the creative name *index.html*

And here's the code:

```html
<!DOCTYPE HTML>
<html>
<head>

    <script type="text/javascript">
         function myWebsocketStart()
         {
               var ws = new WebSocket("ws://localhost:3000/websocket");

               ws.onopen = function()
               {
                  // Web Socket is connected, send data using send()
                  ws.send("ping");
                  var myTextArea = document.getElementById("textarea1");
                  myTextArea.value = myTextArea.value + "\n" + "First message sent";
               };

               ws.onmessage = function (evt)
               {
                  var myTextArea = document.getElementById("textarea1");
                  myTextArea.value = myTextArea.value + "\n" + evt.data
                  if(evt.data == "pong") {
                    setTimeout(function(){ws.send("ping");}, 2000);
                  }
               };

               ws.onclose = function()
               {
                  var myTextArea = document.getElementById("textarea1");
                  myTextArea.value = myTextArea.value + "\n" + "Connection closed";
               };

         }

    </script>

</head>
<body>
<button onclick="javascript:myWebsocketStart()">Start websocket!</button>
<textarea id="textarea1">MyTextArea</textarea>
</body>
</html>
```

The only differences are, that when we open the connection, we send a "ping" message and notify our user about it. Now, whenever we get back a "pong" message, we append it to our textarea and after 2 seconds we answer with a "ping" message again.

We will again need the basic go file structure with the upgrader defined already, and the connection created:

```go
package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"fmt"
	"io/ioutil"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func main() {
	indexFile, err := os.Open("html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}
	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r * http.Request) {
		fmt.Fprintf(w, string(index))
	})
	http.ListenAndServe(":3000", nil)
}
```
### Writing the websocket code

Ok, so now, whenever we get a "ping" message, we wait 2 seconds and answer with a "pong" message. If we get anything else, we just close the connection.

```go
conn, err := upgrader.Upgrade(w, r, nil)
if err != nil {
  fmt.Println(err)
  return
}
for {
  msgType, msg, err := conn.ReadMessage()
  if err != nil {
    fmt.Println(err)
    return
  }
  if string(msg) == "ping" {
    fmt.Println("ping")
    time.Sleep(2 * time.Second)
    err = conn.WriteMessage(msgType, []byte("pong"))
    if err != nil {
      fmt.Println(err)
      return
    }
  } else {
    conn.Close()
    fmt.Println(string(msg))
    return
  }
}
```
Using the ReadMessage function on our connection we get the type, content, and maybe error. We check the message and answer.

## Conclusion

That's actually all, have fun with it and build something great!
