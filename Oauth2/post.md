# Getting started with OAuth2 in Go
## Introduction

Authentication is usually a crucial part in any web app. You could always roll your own authentication mechanics if you wanted, however, this creates an additional barrier between the user and your web app: Registration.

That's why OAuth, and earlier OAuth2, was created. It makes it much more convenient to log in to your app, because the user can log in with one of the many accounts he already has.

## What we'll cover in this tutorial

We will set up a web app with OAuth2 provided by Google. For this we'll need to:
1. Create a web app in Google and get our ClientID and a ClientSecret.
2. Put those into accessible and fairly safe places in our system.
3. Plan the structure of our web app.
4. Make sure we have the needed dependencies.
5. Understand how OAuth2 works.
6. Write the application logic.

**Let's begin.**

## Creating a project in Google and getting the client ID and secret

First, go to the [Google Cloud Platform][1] and create a new project. Later open the left menu, and open the ***API Manager***. There, search for the ***Google+ API*** and enable it.

Next, open the credentials submenu, and choose ***Create credentials -> OAuth client ID***. The application type should be ***Web application*** and give it a custom name if you want. In "Authorized JavaScript origins" put in the address of the site you'll be login in from. I will use http://localhost:3000. Then, in the field ***Authorized redirect URLs*** put in the address of the site, to which the user will be redirected after logging in. I'll use http://localhost:3000/GoogleCallback.

Now the ***client ID*** and ***client secret*** should be displayed for you. Write them down somewhere safe. Remember that the client secret has to stay secret for the entire lifetime of your app.

## Safely storing the client ID and secret

There are many ways to safely store the client ID and secret. In production you should make sure that the client secret remains secret.

In this tutorial we won't cover this. Instead, we will store those variables as system environment variables. Now:
* Create an environment variable called ***googlekey*** holding your client ID.
* Create an environment variable called ***googlesecret*** holding your client secret.

## Planning the structure

In this tutorial we'll write code in one file. In production you would want to split this into multiple files.

Let's start with a basic go web app structure:

```go
package main

import (
  "fmt"
  "net/http"
)

func main() {
  fmt.Println(http.ListenAndServe(":3000", nil))
}
```

Now we'll set up a simple site:

```go
const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`
```

We will also need:
* The home page, where we will click the login button from.
* The page handling redirection to the google service.
* The callback page handling the information we get from Google.

So let's set up the base structure for that:

```go
func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/GoogleLogin", handleGoogleLogin)
	http.HandleFunc("/GoogleCallback", handleGoogleCallback)
	fmt.Println(http.ListenAndServe(":3000", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
}
```
## Dependencies

You will need to
```
go get golang.org/x/oauth2
```
if you don't have it already.

## Understanding OAuth2

To really integrate OAuth2 into our web application it's good to understand how it works.
That's the flow of OAuth2:
1. The user opens the website and clicks the login button.
2. The user gets redirected to the google login handler page. This page generates a random state string by which it will identify the user, and constructs a google login link using it. The user then gets redirected to that page.
3. The user logs in and gets back a code and the random string we gave him. He gets redirected back to our page, using a POST request to give us the code and state string.
4. We verify if it's the same state string. If it is then we use the code to ask google for a short-lived ***access token***. We can save the code for future use to get another token later.
5. We use the ***token*** to initiate actions regarding the user account.

## Writing the application logic
Before starting remember to import the *golang.org/x/oauth2* package.
To begin with, let's write the home page handler:
```go
func handleMain(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, htmlIndex)
}
```

Next we need to create a variable we'll use for storing data and communicating with Google and the ***random state variable***:
```go
var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:3000/GoogleCallback",
		ClientID:     os.Getenv("googlekey"),
		ClientSecret: os.Getenv("googlesecret"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
// Some random string, random for each request
	oauthStateString = "random"
)
```

The *Scopes* variable defines the amount of access we get over the users account.

Note that the *oauthStateString* should be randomly generated on a per user basis.

#### Handling communication with Google

This is the code that creates a login link and redirects the user to it:
```go
func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
```

We use the *googleOauthConfig* variable to create a login link using the random state variable, and later redirect the user to it.

---

Now we need the logic that get's the code after the user logs in and checks if the state variable matches:
```go
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
		fmt.Println("Code exchange failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	fmt.Fprintf(w, "Content: %s\n", contents)
}
```

First we check the state variable, and notify the user if it doesn't match. If it matches we get the code and communicate with google using the *Exchange* function. We have no context so we use *NoContext*.

Later, if we successfully get the token we make a request to google passing the token with it and get the users *userinfo*. We print the response to our user.

## Conclusion

That's all we have to do to integrate OAuth2 into our Golang application. I hope that I helped someone with this problems as I really couldn't find beginner-suited, detailed resources about OAuth2 in Go.

*Now go and build something amazing!*

[1]:Dashboardconsole.developers.google.com
