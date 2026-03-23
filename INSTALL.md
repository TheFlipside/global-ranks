# Installation Guide

Production deployment on Debian/Ubuntu with systemd, PostgreSQL, and nginx.

## 1. Install packages

```bash
sudo apt-get update
sudo apt-get install -y golang git postgresql postgresql-contrib nginx certbot python3-certbot-nginx
```

Verify Go is at least version 1.22:

```bash
go version
```

If the distribution ships an older version, install from <https://go.dev/dl/> instead.

## 2. Clone and build

```bash
cd /tmp
git clone https://git.fiedler.live/tux/global-ranks.git
cd global-ranks
make build
```

Install the binary and migration files:

```bash
sudo mkdir -p /opt/global-ranks/migrations
sudo cp bin/global-ranks /opt/global-ranks/
sudo cp migrations/*.sql /opt/global-ranks/migrations/
```

## 3. Create a service user

```bash
sudo useradd --system --no-create-home --shell /usr/sbin/nologin globalranks
```

Set the directory permissions for the service user:

```bash
sudo chown -R globalranks /opt/global-ranks/
```

## 4. Set up PostgreSQL

Switch to the 'postgres' user and create the database and role:

```bash
sudo -u postgres psql
```

```sql
CREATE USER globalranks WITH PASSWORD 'choose-a-strong-password';
CREATE DATABASE globalranks OWNER globalranks;
\q
```

Verify connectivity:

```bash
psql "postgres://globalranks:choose-a-strong-password@localhost/globalranks" -c "SELECT 1;"
```

Schema migrations run automatically when the service starts.
No manual SQL import is needed.

## 5. Configure the service

Create an environment file that systemd will source. Only root should be able
to read it since it contains the database password.

```bash
sudo tee /etc/default/global-ranks > /dev/null << 'EOF'
GR_DB_DSN="postgres://globalranks:choose-a-strong-password@localhost/globalranks?sslmode=disable"
GR_PORT=8080
GR_TRUSTED_PROXIES="127.0.0.1,::1"

# Tune if needed (defaults shown):
# GR_RATE_SCORE_PER_SEC=0.2
# GR_RATE_SCORE_BURST=3
# GR_RATE_GENERAL_PER_SEC=10
# GR_RATE_GENERAL_BURST=30
# GR_MAX_SCORE=999999999
# GR_AVATAR_CACHE_SIZE=1000
EOF

sudo chmod 600 /etc/default/global-ranks
```

## 6. Install the systemd unit

```bash
sudo tee /etc/systemd/system/global-ranks.service > /dev/null << 'EOF'
[Unit]
Description=GlobalRanks Leaderboard Service
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=globalranks
Group=globalranks
EnvironmentFile=/etc/default/global-ranks
ExecStart=/opt/global-ranks/global-ranks
WorkingDirectory=/opt/global-ranks
Restart=on-failure
RestartSec=5

# Hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
PrivateTmp=true
PrivateDevices=true
ReadOnlyPaths=/opt/global-ranks

[Install]
WantedBy=multi-user.target
EOF
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now global-ranks
```

Check that it is running:

```bash
sudo systemctl status global-ranks
curl -s http://localhost:8080/api/v1/health
```

The health endpoint should return `{"status":"ok","db":"ok"}`.

## 7. Configure nginx

Copy the provided configuration and adjust the 'server_name' to your domain:

```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/global-ranks
sudo sed -i 's/ranks.example.com/your-domain.com/g' /etc/nginx/sites-available/global-ranks
sudo ln -s /etc/nginx/sites-available/global-ranks /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
```

Test and reload:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## 8. Obtain a TLS certificate

```bash
sudo certbot --nginx -d your-domain.com
```

Certbot will modify the nginx config to point at the issued certificate and set
up automatic renewal via a systemd timer.

## 9. Verify the deployment

From an external machine:

```bash
# Health check
curl https://your-domain.com/api/v1/health

# Register a test user
curl -X POST https://your-domain.com/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"uuid":"550e8400-e29b-41d4-a716-446655440000"}'
```

## Updating

To deploy a new version:

```bash
cd /tmp/global-ranks
git pull
make build
sudo cp bin/global-ranks /opt/global-ranks/
sudo cp migrations/*.sql /opt/global-ranks/migrations/
sudo systemctl restart global-ranks
```

New migrations are applied automatically on startup.

## Troubleshooting

| Symptom | Check |
|---------|-------|
| Service won't start | 'journalctl -u global-ranks -e' |
| Database connection refused | 'sudo systemctl status postgresql' and verify DSN |
| 502 from nginx | Confirm the service is listening: 'ss -tlnp \| grep 8080' |
| Rate limiter blocks all requests | Verify 'GR_TRUSTED_PROXIES' includes the nginx IP |
