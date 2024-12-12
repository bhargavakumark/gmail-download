package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"google.golang.org/api/gmail/v1"
)

func processEmails(service *gmail.Service, userID string, labelAction LabelAction) {
	for _, action := range labelAction.Actions {
		query := fmt.Sprintf("label:%s subject:%s", labelAction.Label, action.SubjectFilter)
		nextPageToken := ""
		for {
			msgs, err := service.Users.Messages.List(userID).Q(query).PageToken(nextPageToken).Do()
			if err != nil {
				log.Printf("Unable to list messages for label %s: %v", labelAction.Label, err)
				break
			}

			log.Printf("Processing messages for label %s...\n", labelAction.Label)
			for _, msg := range msgs.Messages {
				log.Printf("Processing msg %s...\n", msg.Id)
				m, err := service.Users.Messages.Get(userID, msg.Id).Do()
				if err != nil {
					log.Printf("Unable to retrieve message: %v", err)
					continue
				}

				if action.Download {
					for _, part := range m.Payload.Parts {
						if part.Filename != "" && part.Body.AttachmentId != "" {
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
								dir = "."
							}
							filePath := fmt.Sprintf("%s/%s", dir, part.Filename)
							if err := os.MkdirAll(dir, 0o755); err != nil {
								log.Printf("Failed to create directory %s: %v", dir, err)
								continue
							}
							if err := os.WriteFile(filePath, data, 0o644); err != nil {
								log.Printf("Failed to save attachment: %v", err)
							} else {
								log.Printf("Saved attachment: %s", filePath)
							}
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
