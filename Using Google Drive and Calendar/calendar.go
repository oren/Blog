package main

import (
	"fmt"
	"net/http"
	"golang.org/x/oauth2"
	"os"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
//"io/ioutil"
	"golang.org/x/net/context"
	"time"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:3000/GoogleCallback",
		ClientID:     os.Getenv("googlekey"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientSecret: os.Getenv("googlesecret"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		Scopes:       []string{"https://www.googleapis.com/auth/calendar"},
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

	calendarService, err := calendar.New(client)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

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

	newEvent := calendar.Event{Summary: "Testevent", Start: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 12, 24, 0, 0, time.UTC).Format(time.RFC3339)}, End: &calendar.EventDateTime{DateTime: time.Date(2016, 3, 11, 13, 24, 0, 0, time.UTC).Format(time.RFC3339)}}
	createdEvent, err := calendarService.Events.Insert("primary", &newEvent).Do()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	fmt.Fprintln(w, "New event in your calendar: \"", createdEvent.Summary, "\" starting at ", createdEvent.Start.DateTime)
}
