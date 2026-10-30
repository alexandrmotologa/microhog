# Deploying MicroHog

MicroHog compiles down to a single binary with zero external runtime dependencies. It runs on any Linux, macOS, or Windows host, including small $4 per month virtual private servers.

## Running the Binary

### 1. Download or Build

```bash
git clone https://github.com/alexandrmotologa/microhog.git
cd microhog
go build -o microhog cmd/microhog/main.go
```

### 2. Start the Server

```bash
./microhog serve --port 8080 --db ./data/microhog.db
```

Parameters:
- `--port`: The listening port (default: `8080`).
- `--db`: Path to the SQLite database file (default: `./microhog.db`).
- `--salt-secret`: Optional key for daily visitor hashing. If omitted, an internal secret is generated on startup.
- `--retention-days`: Automatically prune events older than N days (default: `90`).

---

## Docker Deployment

Build and run using Docker:

```bash
docker build -t microhog:latest .
docker run -d \
  --name microhog \
  -p 8080:8080 \
  -v $(pwd)/data:/data \
  microhog:latest serve --port 8080 --db /data/microhog.db
```

---

## Systemd Service (Ubuntu / Debian)

Create a systemd unit file at `/etc/systemd/system/microhog.service`:

```ini
[Unit]
Description=MicroHog Analytics Engine
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/lib/microhog
ExecStart=/usr/local/bin/microhog serve --port 8080 --db /var/lib/microhog/microhog.db
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now microhog
```

---

## Reverse Proxy (Caddy / Nginx)

### Caddy
```caddyfile
analytics.yourdomain.com {
    reverse_proxy localhost:8080
}
```

### Nginx
```nginx
server {
    server_name analytics.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Server-Sent Events (SSE) support
        proxy_set_header Connection '';
        proxy_http_version 1.1;
        chunked_transfer_encoding off;
        proxy_buffering off;
        proxy_cache off;
    }
}
```
