package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/getlantern/systray"
	gwnhttp "github.com/link00000000/google-workspace-notify/src/httpserver"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var ErrExitRequested = errors.New("exit requested by user")

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	stateToken := "state-token" // TODO: Generate proper state token
	authURL := config.AuthCodeURL(stateToken, oauth2.AccessTypeOffline)

	// TODO: Automatically open browser
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	// TODO: Run HTTP server locally to recieve code
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	// TODO: Create a tray icon for managing

	ctx, cancel := context.WithCancelCause(context.Background())

	go func() {
		onReady := func() {
			systray.SetIcon(trayIcon)
			systray.SetTitle("Google Workspace Notify")
			systray.SetTooltip("Settings")

			mSettings := systray.AddMenuItem("Settings", "Open settings in web browser")
			go func() {
				<-mSettings.ClickedCh

				var cmd string
				var args []string

				switch runtime.GOOS {
				case "windows":
					cmd = "explorer.exe"
				case "darwin":
					cmd = "open"
					args = []string{}
				default:
					cmd = "xdg-open"
					args = []string{}
				}

				args = append(args, "http://localhost:8080/settings")

				log.Printf("executing command: %s", strings.Join(append([]string{cmd}, args...), " "))
				err := exec.Command(cmd, args...).Start()
				if err != nil {
					log.Printf("failed to execute command \"%s\": %v", strings.Join(append([]string{cmd}, args...), " "), err)
				}
			}()

			systray.AddSeparator()

			mExit := systray.AddMenuItem("Exit", "Exit")
			go func() {
				<-mExit.ClickedCh
				cancel(ErrExitRequested)
			}()
		}

		onExit := func() {
		}

		systray.Run(onReady, onExit)
	}()
	defer systray.Quit()

	s := gwnhttp.NewGWNHttpServer()
	go func() {
		s.ListenAndServe(ctx)
	}()

	log.Println("waiting for context")
	<-ctx.Done()
	if context.Cause(ctx) != nil && context.Cause(ctx) != ErrExitRequested {
		log.Fatalf("main context completed with error: %v", context.Cause(ctx))
	}

	return

	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	gmailSrv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	r, err := gmailSrv.Users.Labels.List(user).Do()
	if err != nil {
		log.Fatalf("unable to retrieve labels: %v", err)
	}
	if len(r.Labels) == 0 {
		fmt.Println("no labels found.")
		return
	}
	fmt.Println("Labels:")
	for _, l := range r.Labels {
		fmt.Printf("- %s\n", l.Name)
	}

	// TODO: Poll for emails https://developers.google.com/workspace/gmail/api/guides/sync
	// TODO: Queue a desktop notification for new emails

	calSrv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("unable to retrieve Calendar client: %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := calSrv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(10).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}

	// TODO: Poll for reminders
	// TODO: Queue a desktop notification for the reminders
}
