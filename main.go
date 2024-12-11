package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// userID is the user's email address. The special value "me" can be used to indicate the authenticated user.
// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/get
const userID = "me"

var (
	// destinationDir is the path to the directory where the attachments will be saved.
	crendentialJsonPath = os.Getenv("CREDENTIALS_JSON")

	// destinationDir is the path to the directory where the attachments will be saved.
	destinationDir = os.Getenv("DESTINATION_DIR")

	// tokenFile is the path to the file containing the token.
	tokenFile = os.Getenv("CLIENT_TOKEN_FILE")
)

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
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

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
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// getAttachments retrieves the attachments from the messages matching the query
// and saves them to the destination directory.
func getAttachments(ctx context.Context, svc *gmail.Service, query string) {
	pageToken := ""
	for {
		// Listing messages.
		listResp, err := svc.Users.Messages.List(userID).Q(query).PageToken(pageToken).Context(ctx).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve messages: %v", err)
		}

		log.Printf("Processing %d message(s)...\n", len(listResp.Messages))
		for _, msgInfo := range listResp.Messages {
			// Getting a message.
			log.Printf("Processing msg %s...\n", msgInfo.Id)
			msg, err := svc.Users.Messages.Get(userID, msgInfo.Id).Context(ctx).Do()
			if err != nil {
				log.Fatalf("Unable to retrieve message: %v", err)
			}

			for _, part := range msg.Payload.Parts {
				if part.Filename == "" { // Only present for attachments
					continue
				}
				// Getting an attachment.
				att, err := svc.Users.Messages.Attachments.Get(userID, msg.Id, part.Body.AttachmentId).Context(ctx).Do()
				if err != nil {
					log.Fatalf("Unable to retrieve attachment: %v", err)
				}

				// Saving the attachment.
				saveAttachment(att.Data, part.Filename)
			}
		}

		if pageToken = listResp.NextPageToken; pageToken == "" {
			break
		}
	}
}

// saveAttachment saves the attachment to the destination directory.
func saveAttachment(data, filename string) {
	// Decoding the base64url encoded string into bytes.
	b, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		log.Fatalf("Unable to decode data: %v", err)
	}

	// Creating a file for saving the attachment.
	log.Printf("Saving %s...\n", filename)
	out, err := os.Create(destinationDir + filename)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer func() {
		if err = out.Close(); err != nil {
			log.Fatalf("Unable to close the file: %v", err)
		}
	}()

	if _, err = out.Write(b); err != nil {
		log.Fatalf("Unable to write file: %v", err)
	}
	if err = out.Sync(); err != nil {
		log.Fatalf("Unable to sync commit: %v", err)
	}
}

func main() {
	if os.Getenv("CREDENTIALS_JSON") == "" {
		log.Fatalf("Env variable CREDENTIALS_JSON not set")
	}

	ctx := context.Background()

	// Load credentials.json from the same directory as the program
	b, err := os.ReadFile(os.Getenv("CREDENTIALS_JSON"))
	if err != nil {
		log.Fatalf("unable to read client secret file: %v", err)
	}

	// Authenticate with Gmail API
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}

	/*
		// Setting up the config manually. You can also retrieve a config from a file, e.g.
		// config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope)
		// where b corresponds to the contents of the client credentials file.
		config = &oauth2.Config{
			ClientID:     os.Getenv("CLIENT_ID"),
			ClientSecret: os.Getenv("CLIENT_SECRET"),
			Endpoint:     google.Endpoint,
			RedirectURL:  "http://localhost", // Not required once you have a token
			Scopes:       []string{gmail.GmailReadonlyScope},
		}
	*/

	// Retrieving a token for first time and saving it.
	// Required if you don't have a token, or it's expired.
	// saveToken(getTokenFromWeb(ctx, config))

	// Setting up the token manually. You can also retrieve a token from a file, e.g.
	// token := tokenFromFile()
	/*
		token := &oauth2.Token{
			AccessToken:  os.Getenv("ACCESS_TOKEN"),
			TokenType:    "Bearer",
			RefreshToken: os.Getenv("REFRESH_TOKEN"),
			Expiry:       time.Date(2023, time.March, 27, 15, 5, 14, 0, time.UTC),
		}
	*/

	client := getClient(config)

	// Creating the gmail service.
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create gmail service: %v", err)
	}

	log.Fatalf("Force terminating")
	// Getting the attachments.
	query := "from:example@mail.com has:attachment subject:(My Subject)"
	getAttachments(ctx, svc, query)
}
