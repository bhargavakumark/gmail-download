package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
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
// Starts a local server to automatically capture the OAuth redirect.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Set redirect URI to localhost
	redirectURL := "http://localhost:9901/callback"
	config.RedirectURL = redirectURL

	// Channel to receive the authorization code
	codeChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Start local server to handle OAuth callback
	listener, err := net.Listen("tcp", ":9901")
	if err != nil {
		log.Fatalf("Unable to start local server: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errorChan <- fmt.Errorf("no authorization code in callback")
			http.Error(w, "Authorization failed: no code received", http.StatusBadRequest)
			return
		}

		codeChan <- code
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<html>
			<head><title>Authorization Successful</title></head>
			<body>
				<h1>Authorization Successful!</h1>
				<p>You can close this window and return to the application.</p>
			</body>
			</html>
		`)
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errorChan <- fmt.Errorf("server error: %v", err)
		}
		listener.Close()
	}()

	// Generate authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Opening browser for authorization...\n")
	fmt.Printf("If browser doesn't open automatically, go to: %v\n", authURL)

	// Try to open browser automatically (optional)
	openBrowser(authURL)

	// Wait for authorization code or error
	var authCode string
	select {
	case authCode = <-codeChan:
		// Got the code, shutdown server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	case err := <-errorChan:
		server.Shutdown(context.Background())
		log.Fatalf("Authorization error: %v", err)
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		log.Fatalf("Authorization timeout: no response received within 5 minutes")
	}

	// Exchange authorization code for token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// openBrowser opens the URL in Firefox (cross-platform)
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "windows":
		// Try Firefox first, fallback to default browser
		err = exec.Command("cmd", "/c", "start", "firefox", url).Start()
		if err != nil {
			err = exec.Command("cmd", "/c", "start", url).Start()
		}
	case "darwin":
		// macOS: use open -a to launch Firefox specifically
		err = exec.Command("open", "-a", "Firefox", url).Start()
	case "linux":
		// Try Firefox first, fallback to xdg-open
		err = exec.Command("firefox", url).Start()
		if err != nil {
			err = exec.Command("xdg-open", url).Start()
		}
	default:
		err = fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	if err != nil {
		// Non-fatal: user can manually open the URL
		log.Printf("Could not automatically open Firefox: %v", err)
		log.Printf("Please manually open: %v", url)
	}
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
