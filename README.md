# gmail-download

A simple program to download emails from gmail and save them to local folder. And optionally delete them from gmail

## Google cloud project and API credentials

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


## Gmail Token

When you run the program first time, it prompts you to authorize access:

```
Go to the following link in your browser then type the authorization code:`
```

If you're not already signed in to your Google Account, sign in when prompted. If you're signed in to multiple accounts, select one account to use for authorization.

Once authorizaiton you will be redirected to a URL like `http://localhost` whith will fail with 404 not found.

The code you need to copy/paste into your terminal is in the URL. \&code=<long-code-is-here-copy-this>=https://. Copy this code and paste as input to the this program. It will save the token as `token.json` locally. Next time you run the program, you aren't prompted for authorization.

Note: The authentication scope required for doing changes to gmail requires app to be verified as per https://developers.google.com/gmail/api/auth/scopes#scopes.
