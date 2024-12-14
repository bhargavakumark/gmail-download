package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"google.golang.org/api/gmail/v1"
)

type Action struct {
	SubjectFilter   string `json:"subject_filter"`
	Download        bool   `json:"download_attachment"`
	MarkAsRead      bool   `json:"mark_as_read"`
	Delete          bool   `json:"delete_email"`
	SaveTo          string `json:"save_to"`
	PdfPassword     string `json:"pdf_password"`
	FilenamePattern string `json:"filename_pattern"`
	SaveAsPdf       bool   `json:"save_as_pdf"`
}

type LabelAction struct {
	Label   string   `json:"label"`
	Actions []Action `json:"actions"`
}

// saveEmailAsPDF saves the email content as a PDF file.
func saveEmailAsPDF(emailID, emailDate, subject, body, saveDir string) error {
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		return fmt.Errorf("save directory does not exist: %s", saveDir)
	}

	filename := fmt.Sprintf("%s/email_%s_%s.pdf", saveDir, emailDate, emailID)
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	pdf.CellFormat(0, 10, fmt.Sprintf("Email ID: %s", emailID), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Date: %s", emailDate), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 10, fmt.Sprintf("Subject: %s", subject), "", 1, "L", false, 0, "")

	// Add a line break
	pdf.Ln(10)

	pdf.MultiCell(0, 10, body, "", "L", false)

	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %v", err)
	}

	log.Printf("Saved email as PDF: %s", filename)
	return nil
}

func formatFilename(pattern, originalFilename, emailDate string) string {
	// Replace placeholders in the pattern with actual values
	formatted := strings.ReplaceAll(pattern, "{original}", originalFilename)
	formatted = strings.ReplaceAll(formatted, "{date}", emailDate)
	return formatted
}

func parseEmailDate(dateStr string) string {
	// Define possible layouts for parsing email date
	layouts := []string{
		time.RFC1123Z,                    // Example: Mon, 02 Jan 2006 15:04:05 -0700
		time.RFC1123,                     // Example: Mon, 02 Jan 2006 15:04:05 MST
		"Mon, 2 Jan 2006 15:04:05 -0700", // Single-digit day
		"2 Jan 2006 15:04:05 -0700",      // No weekday
	}

	for _, layout := range layouts {
		if parsedTime, err := time.Parse(layout, dateStr); err == nil {
			return parsedTime.Format("2006-01-02_15-04-05")
		}
	}

	// Return "unknown" if parsing fails for all layouts
	log.Printf("ERROR: Failed to parse email date: %s", dateStr)
	return "unknown"
}

func processEmails(service *gmail.Service, userID string, labelAction LabelAction) {
	log.Printf("Processing label: %s", labelAction.Label)
	for _, action := range labelAction.Actions {
		query := fmt.Sprintf("label:%s subject:%s", labelAction.Label, action.SubjectFilter)
		nextPageToken := ""
		for {
			msgs, err := service.Users.Messages.List(userID).Q(query).PageToken(nextPageToken).Do()
			if err != nil {
				log.Printf("Unable to list messages for label %s: %v", labelAction.Label, err)
				break
			}

			for _, msg := range msgs.Messages {
				m, err := service.Users.Messages.Get(userID, msg.Id).Do()
				if err != nil {
					log.Printf("Unable to retrieve message: %v", err)
					continue
				}

				// Parse email date/time
				emailDate := "unknown"
				for _, header := range m.Payload.Headers {
					if header.Name == "Date" {
						emailDate = parseEmailDate(header.Value)
						break
					}
				}

				if action.Download {
					for _, part := range m.Payload.Parts {
						if part.Filename == "" || part.Body.AttachmentId == "" {
							continue
						}

						attachment, err := service.Users.Messages.Attachments.Get(userID, msg.Id, part.Body.AttachmentId).Do()
						if err != nil {
							log.Printf("Unable to retrieve attachment: %v", err)
							continue
						}

						data, err := base64.URLEncoding.DecodeString(attachment.Data)
						if err != nil {
							log.Printf("Failed to decode attachment data: %v", err)
							continue
						}

						dir := action.SaveTo
						if dir == "" {
							log.Fatalf("SaveTo directory is empty for action: %+v", action)
						}
						if _, err := os.Stat(dir); os.IsNotExist(err) {
							log.Fatalf("SaveTo directory does not exist: %s", dir)
						}

						// Apply filename pattern
						filename := part.Filename
						if action.FilenamePattern != "" {
							filename = formatFilename(action.FilenamePattern, part.Filename, emailDate)
						}

						filePath := fmt.Sprintf("%s/%s", dir, filename)
						if err := os.WriteFile(filePath, data, 0644); err != nil {
							log.Printf("Failed to save attachment: %v", err)
							continue
						}
						log.Printf("Saved attachment: %s", filePath)

						if action.PdfPassword != "" && part.Filename[len(part.Filename)-4:] == ".pdf" {
							c := model.NewDefaultConfiguration()
							c.UserPW = action.PdfPassword
							c.Cmd = model.DECRYPT
							err := api.DecryptFile(filePath, filePath, c)
							if err != nil {
								log.Fatalf("Failed to decrypt PDF file %s: %v", filePath, err)
							}
							log.Printf("Successfully decrypted PDF: %s", filePath)
						}
					}
				}

				if action.SaveAsPdf {
					emailDate := "unknown"
					for _, header := range m.Payload.Headers {
						if header.Name == "Date" {
							emailDate = parseEmailDate(header.Value)
							break
						}
					}

					// Extract subject
					subject := "No Subject"
					for _, header := range m.Payload.Headers {
						if header.Name == "Subject" {
							subject = header.Value
							break
						}
					}

					// Extract body
					body := ""
					if m.Payload.Body != nil && m.Payload.Body.Data != "" {
						data, err := base64.URLEncoding.DecodeString(m.Payload.Body.Data)
						if err == nil {
							body = string(data)
						}

						err = saveEmailAsPDF(msg.Id, emailDate, subject, body, action.SaveTo)
						if err != nil {
							log.Printf("Failed to save email as PDF: %v", err)
						}
					}
				}

				if action.MarkAsRead {
					_, err := service.Users.Messages.Modify(userID, msg.Id, &gmail.ModifyMessageRequest{
						RemoveLabelIds: []string{"UNREAD"},
					}).Do()
					if err != nil {
						log.Printf("Failed to mark email as read: %v", err)
					}
				}

				if action.Delete {
					log.Printf("Deleting email with ID: %s", msg.Id)
					if err := service.Users.Messages.Delete(userID, msg.Id).Do(); err != nil {
						log.Printf("Failed to delete email: %v", err)
					}
				}

			}

			nextPageToken = msgs.NextPageToken
			if nextPageToken == "" {
				break
			}
		}
	}
}
