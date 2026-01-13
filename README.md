# ict-school-web

This is a Go web application for the "Mock Class Reservation System". It uses **Nix** for reproducible builds and **Docker** for deployment.

## 0. Setup & Installation

Before running the project, you need to install the necessary tools.

### A. On Your Local Machine (The Builder)
You only need **Nix**. You do *not* need Go, Postgres, or Docker installed locally to build the project.

**1. Install Nix:**
Mac / Linux:
```bash
sh <(curl -L [https://nixos.org/nix/install](https://nixos.org/nix/install)) --daemon

```

**2. Enable Flakes (Crucial):**
Nix Flakes are an experimental feature. You must enable them.

```bash
mkdir -p ~/.config/nix
echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf

```

*Restart your terminal after installation.*

---

### B. On the Server (The Runner)

The server (e.g., AWS EC2 Amazon Linux 2023) needs **Docker** and **Docker Compose** to run the image we build.

**1. Install Docker:**

```bash
sudo yum update -y
sudo yum install -y docker
sudo service docker start
sudo usermod -a -G docker ec2-user

```

*(Log out and log back in for the user permission to take effect)*

**2. Install Docker Compose:**

```bash
mkdir -p /usr/libexec/docker/cli-plugins/
curl -SL [https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64](https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64) -o /usr/libexec/docker/cli-plugins/docker-compose
chmod +x /usr/libexec/docker/cli-plugins/docker-compose

```

**3. Verify:**

```bash
docker compose version
# Should see: Docker Compose version v2.x.x

```

---

## 1. Local Development

To modify the code and test locally:

1. **Enter the Dev Shell:**
This downloads Go, Postgres, and sets up environment variables automatically.
```bash
nix develop

```


2. **Run the Server:**
```bash
go run ./cmd/server/main.go

```


The server will start at `http://localhost:8080`.

---

## 2. Building for Production

We use Nix to build a Docker image that is byte-for-byte reproducible.

1. **Build the Image:**
```bash
nix build .#docker

```


This creates a symlink named `result` in your folder. This is a compressed archive of the Docker image.

---

## 3. Deployment Guide

### Step 1: Copy Files to Server

Assuming you have your AWS key (`key.pem`) and the server IP (`1.2.3.4`).

```bash
# 1. Copy the Docker image artifact
scp -i key.pem result ec2-user@1.2.3.4:/home/ec2-user/ict-web.tar.gz

# 2. Copy the configuration files
scp -i key.pem docker-compose.yml init.sql ec2-user@1.2.3.4:/home/ec2-user/

```

### Step 2: Load & Run on Server

SSH into the server:

```bash
ssh -i key.pem ec2-user@1.2.3.4

```

Then run:

```bash
# 1. Load the image into Docker
docker load < ict-web.tar.gz

# 2. Start the application
# (This starts Postgres, runs init.sql, and starts the Web App)
docker compose up -d

```

### Step 3: Check Status

```bash
docker compose ps

```

You should see both containers with Status `Up`.

---

## Troubleshooting

**Q: "Permission denied" when SSHing?**
Make sure your key file is secure.

```bash
chmod 400 key.pem

```

**Q: Database connection refused?**
Check `docker-compose.yml`. The `DATABASE_URL` must point to the service name `db`, not `localhost`.

```yaml
DATABASE_URL: postgres://postgres:pass@db:5432/ict?sslmode=disable

```

**Q: I updated the code, how do I redeploy?**

1. Run `nix build .#docker` locally.
2. `scp` the new `result` to the server.
3. On server: `docker load < ict-web.tar.gz`.
4. On server: `docker compose up -d` (Docker detects the new image and restarts only the web app).

```

```
