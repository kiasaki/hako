# hako

_A box to fill with your words, pictures and documents._

## intro

Hako is a simple web application that offers an interface for organizing and editing plain-text
files in a folder structure. The files themselves are simply saved as is in _Google Cloud Storage_.
Hako also let's you upload raw files (they don't need to be plain-text) so it can be used as a light
weight Dropbox with no syncing but a web UI that allows editing.

## features

**User**

- Organize files in a folder hierarchy, like on your computer.
- View some files right in the web UI: text, markdown, code, images, csv.
- Edit text files right from the web UI for note taking and writing.
- Upload any type of file for archival.
- Make any file public for sharing (Not yet implemented)

**Technical**

- Multi-tenant
- Simplistic UI (no getting lost in features you don't use, but limiting for some use cases)
- Email-only login (no credentials to remember, no signup process)
- No database (only data is files in Google Cloud Storage)
- Fast (if compared to slower webservers from intepreted languages)
- HTML only, no need for JavaScript to be enabled

## getting started

Before you can build and start hako's webserver you'll need to create an `.env` file
and fill in Google Storage and Sendgrid credentials.

It should look like:

`.env`
```
export GOOGLE_APPLICATION_CREDENTIALS='{"type":"service_account","project_id": ...}'
export GOOGLE_BUCKET_ID='bucket-name-here'
export SENDGRID_API_KEY='...'
export APP_JWT_SECRET='something-secret'
export APP_BASE_URL='http://localhost:3000'
```

Then simply running `make` should build `hako` and start it:

```
make
```

## contributing

Contributions are welcomed in the form of issues or pull-requests. If you want to do substantial
changes to fit your way of working with notes / files feel free to fork this repo and work on a
spin off.

## license

MIT. See `LICENSE` file.
