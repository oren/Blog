# Practical Golang: Using Google Drive and Calendar

## Introduction

Integrating Google services into your app can lead to a lot of nice features for your users, and can create a seamless experience for the user. In this tutorial we'll learn how to use the most useful functionalities of ***Google Calendar*** and ***Google Drive***.

## The theory

To begin with, we should understand the methodology of using the ***Google API*** in Golang. For most of their API's I've skimmed through it works like that:

1. Create an **OAuth2 client** from the OAuth2 *access token*.
2. Use the client to create an app service, this will be our interface we'll use to communicate with Google services.
3. We create a request object and set the needed parameters.
4. We start the action, usually using the *Do()* function on the request object.

After you learn it for one Google service, it will be trivial for you to use it for any other.

## Dependencies

Here we will need the Google Calendar and Google Drive libraries, both of which we need version 3. (the newest at the moment)

So make sure to:

```
go get google.golang.org/api/calendar/v3
go get google.golang.org/api/drive/v3
```

## The basic structure

For both of our apps we'll need the same basic OAuth2 app structure. You can learn more about it in [my previous article][1]

```go
package main

import (
	"fmt"
	"net/http"
	"golang.org/x/oauth2"
	"os"
	"golang.org/x/oauth2/google"
	"golang.org/x/net/context"
	"time"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:3000/GoogleCallback",
		ClientID:     os.Getenv("googlekey"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientSecret: os.Getenv("googlesecret"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		Scopes:       []string{},
		Endpoint:     google.Endpoint,
	}
// Some random string, random for each request
	oauthStateString = "random"
)

const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/GoogleLogin", handleGoogleLogin)
	http.HandleFunc("/GoogleCallback", handleGoogleCallback)
	fmt.Println(http.ListenAndServe(":3000", nil))
}
func handleMain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, htmlIndex)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != oauthStateString {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthStateString, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
}
```

There is one thing I haven't covered in my previous blog post, namely the last line:
```go
client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
```

We need an ***OAuth2 client*** to use the Google API, so we create one. It takes a context, for lack of which we just use the background context. It also needs a token source. As we only want to make one request and know that this token will suffice we create a static token source which will always generate the same token which we've passed to it.

## Creating the Calendar app.

### Preperation
First, as described in [my previous article][1], you should enable the Google Calendar API in the ***Google Cloud Console*** for you app.

Also, we'll need to ask for permission, so add the ***https://www.googleapis.com/auth/calendar*** scope to our googleOauthConfig:

```go
googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:3000/GoogleCallback",
		ClientID:     os.Getenv("googlekey"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientSecret: os.Getenv("googlesecret"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
		Endpoint:     google.Endpoint,
	}
```

### The main code

We'll add everything we write right after creating our OAuth2 client.

First, as I described before, we'll need an app service, here it will be the calendar service, so let's create it!

```go
client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))

calendarService, err := calendar.New(client)
if err != nil {
  fmt.Fprintln(w, err)
  return
}
```

It just uses the OAuth client to create the service and errors out if something goes wrong.

#### Listing events

Now we will create a request, add a few optional parameters to it and start it. We'll build it up step by step.
```go
calendarService, err := calendar.New(client)
if err != nil {
  fmt.Fprintln(w, err)
  return
}

calendarService.Events.List("primary")
```

This creates a request to list all events in your primary calendar. You could also name a specific calendar, but using *primary* will take the primary calendar of that user.

So... I think we don't really care about the events 5 years ago. So let's only take the upcoming ones.

```go
calendarService.Events.List("primary").TimeMin(time.Now().Format(time.RFC3339))
```

We add an option *TimeMin* which takes a *DateTime* by string... No idea why it isn't a nice struct like *time.DateTime*. You also need to format it as a string in the RFC3339 format.

Ok... but that could be a lot of events, so we'll just take the 5 first:

```go
calendarService.Events.List("primary").TimeMin(time.Now().Format(time.RFC3339)).MaxResults(5)
```

Now we just have to ***Do()*** it, and store the results:

```go
calendarEvents, err := calendarService.Events.List("primary").TimeMin(time.Now().Format(time.RFC3339)).MaxResults(5).Do()
if err != nil {
  fmt.Fprintln(w, err)
  return
}
```

How can we now do something with the results? Simple! :

```go
calendarEvents, err := calendarService.Events.List("primary").TimeMin(time.Now().Format(time.RFC3339)).MaxResults(5).Do()
if err != nil {
  fmt.Fprintln(w, err)
  return
}

if len(calendarEvents.Items) > 0 {
	for _, i := range calendarEvents.Items {
		fmt.Fprintln(w, i.Summary, " ", i.Start.DateTime)
	}
}
```

We access a list of events using the ***Items*** field in the *calendarEvents* variable, if there is at least one element, then for each element we write the *summary* and *start time* to the *response writer* using a *for range* loop.

#### Creating an event

Ok, we already know how to list events, now let's create an event!
First, we need an event object:

```go
if len(calendarEvents.Items) > 0 {
	for _, i := range calendarEvents.Items {
		fmt.Fprintln(w, i.Summary, " ", i.Start.DateTime)
	}
}
newEvent := calendar.Event{
	Summary: "Testevent",
	Start: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 12, 24, 0, 0, time.UTC).Format(time.RFC3339)},
	End: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 13, 24, 0, 0, time.UTC).Format(time.RFC3339)},
}
```

We create an Event struct and pass in the ***summary*** - title of the event.
We also pass the start and finish ***DateTime***. We create a *date* using the stdlib *time* package, and then convert it to a string in the RFC3339 format. There are tons of other optional fields you can specify if you want to.

Next we need to create an ***insert*** request object:
```go
newEvent := calendar.Event{
	Summary: "Testevent",
	Start: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 12, 24, 0, 0, time.UTC).Format(time.RFC3339)},
	End: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 13, 24, 0, 0, time.UTC).Format(time.RFC3339)},
}
calendarService.Events.Insert("primary", &newEvent)
```

The ***Insert*** request takes two arguments, the calendar name and an event object.

As usual we neeed to ***Do()*** the request! and saving the resulting created event can also come handy in the future:

```go
createdEvent, err := calendarService.Events.Insert("primary", &newEvent).Do()
if err != nil {
	fmt.Fprintln(w, err)
	return
}
```

In the end let's just print some kind of confirmation to the user:

```go
fmt.Fprintln(w, "New event in your calendar: \"", createdEvent.Summary, "\" starting at ", createdEvent.Start.DateTime)
```

### Hint

You can define the event ID yourself before creating it, but you can also let the Google Calendar service generate an ID for us as we did.

## Creating the Drive app




[1]:https://jacobmartins.com/2016/02/29/getting-started-with-oauth2-in-go/
