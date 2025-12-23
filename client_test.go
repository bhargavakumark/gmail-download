package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenFromFile(t *testing.T) {
	// Create a temporary file with valid token data
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "token.json")

	// Create a valid token
	expectedToken := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Write token to file
	data, err := json.Marshal(expectedToken)
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}
	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Test reading the token
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		t.Fatalf("tokenFromFile() error = %v, want nil", err)
	}

	if token.AccessToken != expectedToken.AccessToken {
		t.Errorf("tokenFromFile() AccessToken = %v, want %v", token.AccessToken, expectedToken.AccessToken)
	}
	if token.RefreshToken != expectedToken.RefreshToken {
		t.Errorf("tokenFromFile() RefreshToken = %v, want %v", token.RefreshToken, expectedToken.RefreshToken)
	}
}

func TestTokenFromFile_NotFound(t *testing.T) {
	// Test with non-existent file
	token, err := tokenFromFile("nonexistent.json")
	if err == nil {
		t.Error("tokenFromFile() error = nil, want error")
	}
	if token != nil {
		t.Errorf("tokenFromFile() token = %v, want nil", token)
	}
}

func TestTokenFromFile_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(tokenFile, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write invalid token file: %v", err)
	}

	// Test reading invalid token
	token, err := tokenFromFile(tokenFile)
	if err == nil {
		t.Error("tokenFromFile() error = nil, want error for invalid JSON")
	}
	// Note: tokenFromFile returns a token struct even on decode error (with zero values)
	// This is the actual behavior of the function
	if token == nil {
		t.Error("tokenFromFile() token = nil, but function returns non-nil token even on decode error")
	}
}

func TestSaveToken(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "saved_token.json")

	token := &oauth2.Token{
		AccessToken:  "test-access-token",
		TokenType:    "Bearer",
		RefreshToken: "test-refresh-token",
		Expiry:       time.Now().Add(1 * time.Hour),
	}

	// Save token
	saveToken(tokenFile, token)

	// Verify file exists
	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		t.Fatalf("saveToken() file was not created: %v", err)
	}

	// Read and verify token
	savedToken, err := tokenFromFile(tokenFile)
	if err != nil {
		t.Fatalf("Failed to read saved token: %v", err)
	}

	if savedToken.AccessToken != token.AccessToken {
		t.Errorf("saveToken() AccessToken = %v, want %v", savedToken.AccessToken, token.AccessToken)
	}
	if savedToken.RefreshToken != token.RefreshToken {
		t.Errorf("saveToken() RefreshToken = %v, want %v", savedToken.RefreshToken, token.RefreshToken)
	}
}

func TestSaveToken_OverwriteExisting(t *testing.T) {
	// Create a temporary file with existing token
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, "existing_token.json")

	oldToken := &oauth2.Token{
		AccessToken: "old-token",
	}
	saveToken(tokenFile, oldToken)

	// Save new token
	newToken := &oauth2.Token{
		AccessToken: "new-token",
	}
	saveToken(tokenFile, newToken)

	// Verify new token was saved
	savedToken, err := tokenFromFile(tokenFile)
	if err != nil {
		t.Fatalf("Failed to read saved token: %v", err)
	}

	if savedToken.AccessToken != newToken.AccessToken {
		t.Errorf("saveToken() did not overwrite existing token: got %v, want %v", savedToken.AccessToken, newToken.AccessToken)
	}
}

