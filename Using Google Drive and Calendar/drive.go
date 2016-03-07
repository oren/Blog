package main

import (
	"fmt"
	"net/http"
	"golang.org/x/oauth2"
	"os"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
//"io/ioutil"
	"golang.org/x/net/context"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:	"http://localhost:3000/GoogleCallback",
		ClientID:     os.Getenv("googlekey"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		ClientSecret: os.Getenv("googlesecret"), // from https://console.developers.google.com/project/<your-project-id>/apiui/credential
		Scopes:       []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/drive.file"},
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

	driveService, err := drive.New(client)


	myFilesList, err := driveService.Files.List().Do()
	if err != nil {
		fmt.Fprintf(w, "Couldn't retrieve files ", err)
	}
	if len(myFilesList.Files) > 0 {
		for _, i := range myFilesList.Files {
			fmt.Fprintln(w, i.Name, " ", i.Id)
		}
	} else {
		fmt.Fprintln(w, "No files found.")
	}
	for myFilesList.NextPageToken != "" {
		myFilesList, err = driveService.Files.List().PageToken(myFilesList.NextPageToken).Do()
		if err != nil {
			fmt.Fprintf(w, "Couldn't retrieve files ", err)
			break
		}
		fmt.Fprintln(w, "Next Page: !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
		if len(myFilesList.Files) > 0 {
			for _, i := range myFilesList.Files {
				fmt.Fprintln(w, i.Name, " ", i.Id)
			}
		} else {
			fmt.Fprintln(w, "No files found.")
		}
	}


	myFile := drive.File{Name: "cats.png"}
	createdFile, err := driveService.Files.Create(&myFile).Do()
	if err != nil {
		fmt.Fprintf(w, "Couldn't create file ", err)
	}
	myImage, err := os.Open("/tmp/image.png")
	if err != nil {
		fmt.Fprintln(w, err)
	}
	updatedFile := drive.File{Name: "catsUpdated.png"}
	_, err = driveService.Files.Update(createdFile.Id, &updatedFile).Media(myImage).Do()
	if err != nil {
		fmt.Fprintln(w, err)
	}
	fmt.Fprintln(w, createdFile.Id)
}
