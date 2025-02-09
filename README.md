# Trisend

Trisend is an SSH-based file transfer solution that simplifies secure file sharing with an HTTP interface for managing access. It enables users to register their public SSH keys via a web interface, facilitating seamless file transfers through SSH commands.

![App Screenshot](./github/app_image.jpg)


## How it works

- **User Registration** – Users register their public SSH keys via the HTTP interface.

- **File Upload** – Authenticated users upload files through SSH.

- **Download Link Creation** – The SSH server generates a secure download link and the session is kept open. When download starts it closes the session.

- **File Retrieval** – The recipient accesses the link to initiate the download.


## Run Locally

Make sure you have golang, node and docker installed.

**Clone the repository**
```bash
  git clone https://github.com/Trijavinc/Trisend.git
```

**Go to the project directory**

```bash
  cd Trisend
```

**build the docker image**

```bash
  make docker
```

**Set up env file**

```bash
PORT=3000
HOST=localhost:3000
SSH_PORT=2222
APP_ENV=dev

# OAuth secrets 
CLIENT_ID=<provide client id>
CLIENT_SECRET=<provide client secret>

SESSION_SECRET=<some-secret>
JWT_SECRET=<some-secret>

# Mail
SMTP_PASSWORD="<provide gmail app password>"
SMTP_USERNAME="<provide an email>"

# Database
DB_PORT=6379
DB_NAME=0
DB_HOST=127.0.0.1
DB_PASSWORD=1234
```

**Start the server**

```bash
  docker compose up
```

## Tech Stack

**Client:** htmx, TailwindCSS

**Server:** Golang, Redis
