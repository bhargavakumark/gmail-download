package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

type Config struct {
	LabelActions []LabelAction `json:"label_actions"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func main() {
	if os.Getenv("GMAIL_CREDENTIALS_JSON") == "" {
		log.Fatalf("Env variable GMAIL_CREDENTIALS_JSON not set")
	}

	userID := os.Getenv("GMAIL_USER")
	if userID == "" {
		log.Fatalf("Env variable GMAIL_USER not set")
	}

	actionFile := os.Getenv("GMAIL_ACTION_CONFIG")
	if actionFile == "" {
		log.Fatalf("Env variable GMAIL_ACTION_CONFIG not set")
	}

	ctx := context.Background()

	actionConfig, err := loadConfig(actionFile)
	if err != nil {
		log.Fatalf("Unable to load config file: %v", err)
	}
	// Check if any action has Delete set to true
	hasDelete := false
	hasModify := false
	for _, labelAction := range actionConfig.LabelActions {
		for _, action := range labelAction.Actions {
			if action.Delete {
				hasDelete = true
				break
			}
			if action.MarkAsRead {
				hasModify = true
				break
			}
		}
	}
	log.Printf("hasDelete: %v, hasModify: %v", hasDelete, hasModify)

	// Load credentials.json from the same directory as the program
	b, err := os.ReadFile(os.Getenv("GMAIL_CREDENTIALS_JSON"))
	if err != nil {
		log.Fatalf("unable to read client secret file: %v", err)
	}

	// Authenticate with Gmail API
	scope := gmail.GmailReadonlyScope
	if hasModify {
		scope = gmail.GmailModifyScope
	}
	if hasDelete {
		scope = gmail.MailGoogleComScope
	}
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("unable to parse client secret file to config: %v", err)
	}

	client := getClient(config)

	// Creating the gmail service.
	svc, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create gmail service: %v", err)
	}

	for _, labelAction := range actionConfig.LabelActions {
		processEmails(svc, userID, labelAction)
	}
}
