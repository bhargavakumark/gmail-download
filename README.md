# Gmail Automation Tool

## Overview

This tool automates various tasks for Gmail accounts using the Gmail API. It allows users to configure actions for specific email labels, such as downloading attachments, marking emails as read, saving emails as PDFs, or deleting emails. The configuration is defined via a JSON file, enabling flexibility and customization.

A simple program to download emails from gmail and save them to local folder. And optionally delete them from gmail

## Features

- **Filter Emails by Label and Subject**: Process emails based on specific Gmail labels and subject filters.
- **Download Attachments**: Save email attachments to a specified directory.
- **Save Emails as PDFs**: Save email content as PDF files with unique filenames.
- **Mark Emails as Read**: Automatically mark processed emails as read.
- **Delete Emails**: Remove emails from the inbox.
- **Secure PDF Processing**: Decrypt PDFs using a provided password.
- **Customizable Filename Patterns**: Rename downloaded files based on email date and a configurable pattern.a

## Prerequisites

Go Environment: Ensure Go is installed and configured.

### Google cloud project and API credentials

Gmail API access is bound to a google cloud project. You can use any existing google cloud project tied to your gmail.com address, or create a new project. 

1. Visit `https://console.cloud.google.com` with your gmail identity and create a new project or use an existing project.
2. Search for `APIs and Services` and visit `Enable APIs and Services` and enable `Gmail API` service. This allows API calls to gmail service on this project. This grants API acesss, not access to actual gmail account.
3. As per google cloud documentation here [Create Credentials](https://developers.google.com/workspace/guides/create-credentials#choose_the_access_credential_that_is_right_for_you), you would need to use OAuth based login to request access to email.
4. In console search for `OAuth consent screen`, you will be provided an option of `User Type` as `Internal` or `External`. If you are not a workspace user, you will not have an option to use `Internal`. Select `External`.
  1. Give some dummy value for `App name` like `gmail-download`.
  2. Give `User support email` as you personal email address.
  3. Give `Developer contact information` as your personal email address.
  4. Click `Save and Continue` to next page.
  5. Click `Save and Continue` in the scopes page.
  6. In `Test Users` screen, add a new user with your personal email address. `Save and Continue`.
5. On the left pane of `API and Services`, select `Credentials`.
  1. Click `Create Credentials`, and select `OAuth Client ID`.
  2. Choose `Desktop App`.
6. Downlaod the credentials and save as `credentials.json`


### Gmail Token

When you run the program first time, it prompts you to authorize access:

```
Go to the following link in your browser then type the authorization code:`
```

If you're not already signed in to your Google Account, sign in when prompted. If you're signed in to multiple accounts, select one account to use for authorization.

Once authorizaiton you will be redirected to a URL like `http://localhost` whith will fail with 404 not found.

The code you need to copy/paste into your terminal is in the URL. \&code=<long-code-is-here-copy-this>=https://. Copy this code and paste as input to the this program. It will save the token as `token.json` locally. Next time you run the program, you aren't prompted for authorization.

Note: The authentication scope required for doing changes to gmail requires app to be verified as per https://developers.google.com/gmail/api/auth/scopes#scopes.

### Environment Variables:

* `GMAIL_CREDENTIALS_JSON`: Path to the credentials.json file.
* `GMAIL_USER`: Gmail user ID (usually your email address).
* `GMAIL_ACTION_CONFIG`: Path to the JSON configuration file.

## Installation

Clone the repository:

```bash
git clone https://github.com/bhargavakumark/gmail-download
cd gmail-download
```

Build the executable:

```bash
go build -o gmail-downlaod
```

## Configuration

Create a JSON configuration file (e.g., config.json) to define actions for specific labels. Here's an example:

```
{
  "label_actions": [
    {
      "label": "INBOX",
      "actions": [
        {
          "subject_filter": "Invoice",
          "download_attachment": true,
          "mark_as_read": true,
          "delete_email": false,
          "save_to": "/path/to/save",
          "pdf_password": "yourpassword",
          "filename_pattern": "attachment_{{date}}_{{email_id}}",
          "save_as_pdf": true
        }
      ]
    }
  ]
}
```

### Configuration Fields

* **label**: Gmail label to filter emails (e.g., "INBOX" or custom labels).
* **subject_filter**: A string to filter emails by subject.
* **download_attachment**: Whether to download attachments (true/false).
* **mark_as_read**: Mark the email as read after processing (true/false).
* **delete_email**: Delete the email after processing (true/false).
* **save_to**: Directory to save downloaded files or PDFs.
* **pdf_password**: Password to decrypt PDFs (leave empty if not needed).
* **filename_pattern**: Pattern for naming files (supports {date} and {email_id} placeholders).
* **save_as_pdf**: Save the email content as a PDF (true/false).

## Usage

Set up the environment variables:

```bash
export GMAIL_CREDENTIALS_JSON=/path/to/credentials.json
export GMAIL_USER=your-email@gmail.com
export GMAIL_ACTION_CONFIG=/path/to/config.json
```

Run the tool:

```bash
./gmail-download
```

## OAuth Scopes

The tool dynamically selects the Gmail API scopes based on the actions specified in the configuration:

* Read-only: https://www.googleapis.com/auth/gmail.readonly (default).
* Modify: https://www.googleapis.com/auth/gmail.modify (for marking as read).
* Full Access: https://mail.google.com/ (for deleting emails).

