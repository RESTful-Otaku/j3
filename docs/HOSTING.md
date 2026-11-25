# Free production & staging hosting

## The short answer

Use an **Oracle Cloud Always Free Ampere A1 VM**. It is the only genuinely free-forever option that
gives you a real always-on Linux server, a persistent disk for the SQLite file, a runnable Go
binary, long-lived SSE and your own HTTPS domains. Render/Fly/Railway/Koyeb/Cloud Run either sleep,
have ephemeral filesystems, cap free usage, or cannot run this Go binary without a rewrite.
Cloudflare Pages is excellent for the static client and downloadable APK/IPA files, but its D1
database is not a mounted SQLite file.

Always Free A1 gives **2 OCPUs / 12 GB RAM**, **200 GB** total storage and **10 TB/month egress** for
one home region. You need an Oracle account (card verification is usual, not charged while free).
A1 capacity can be unavailable in a region; retry or choose another availability domain.

## What to deploy

Release artifacts already include:

- `j3d-linux-arm64` — the headless Go server for the Oracle ARM VM (`j3d-linux-amd64` for x86).
- `web/` — the production SvelteKit static client served by Caddy, including
  `web/openapi.yaml` (live API contract/reference) and `web/docs/` (organiser, runner, official,
  participant, CI, hosting and developer guides).
- `j3-android-qa.apk`, `j3-ios-simulator.zip` (`J3.app`) and, when Apple secrets are configured,
  `j3-ios-qa.ipa` — installable QA apps for devices.
- `LICENSE` — the PolyForm Noncommercial notice.

## Build and deploy from the TUI

`bun run dev` → **Build & deploy ▸** gives a near-one-click path:

1. **Build release (server + web + API + docs)** — regenerates the OpenAPI contract and docs index,
   builds the production web client (with the API reference and guide site), cross-compiles all
   server binaries and writes `dist/`.
2. **Deploy to staging** / **Deploy to production** — packages `dist/` and uploads it using
   `scripts/deploy.sh`.
3. **Build + deploy staging/production (one step)** — does both steps above in sequence.

Set the target once in your shell before launching the TUI (or in your shell profile):

```bash
export DEPLOY_HOST=j3.example.com
export DEPLOY_USER=j3
export DEPLOY_PATH=/srv/j3          # deploy.sh adds prod/ or staging/
export DEPLOY_SSH_KEY=~/.ssh/j3_ed25519
```

The TUI **Docs & developer guides ▸ ☁️ Free hosting plan** opens this page in the documentation
site, and **🔌 API reference** opens the API explorer shipped inside `web/`. To make an APK that
talks to the hosted API, use **Mobile ▸ Build APK for hosted HTTPS server** and enter
`https://api.example.com`.

Run **two isolated instances on the one VM**: production on `127.0.0.1:8787`, staging/QA on
`127.0.0.1:8788`, each with its own SQLite file, env file, CORS origins and subdomain.

## One-time VM setup

```bash
# Ubuntu 24.04 ARM64, shape VM.Standard.A1.Flex (2 OCPU / 12 GB)
sudo apt update && sudo apt install -y caddy sqlite3 ufw
sudo ufw allow 22/tcp && sudo ufw allow 80/tcp && sudo ufw allow 443/tcp && sudo ufw enable
sudo useradd --system --home /var/lib/j3 --shell /usr/sbin/nologin j3
sudo mkdir -p /opt/j3/bin /srv/j3/{downloads,prod,staging} /var/lib/j3/{prod,staging} /var/backups/j3
sudo chown -R j3:j3 /srv/j3 /var/lib/j3 /var/backups/j3
```

Build/cross-compile locally and upload from a release:

```bash
bash scripts/release.sh
scp build/server/j3d-linux-arm64 user@vm:/tmp/j3d
ssh user@vm 'sudo install -m 0755 /tmp/j3d /opt/j3/bin/j3d.new && sudo mv /opt/j3/bin/j3d.new /opt/j3/bin/j3d'
```

## systemd units

`/etc/systemd/system/j3d-prod.service` (copy to `j3d-staging.service`, changing the name, port,
paths and env file):

```ini
[Unit]
Description=J3 venue server (prod)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=j3
Group=j3
WorkingDirectory=/var/lib/j3/prod
EnvironmentFile=/etc/j3/prod.env
ExecStart=/opt/j3/bin/j3d -addr 127.0.0.1:8787 -db /var/lib/j3/prod/j3.db
Restart=always
RestartSec=2
UMask=0077
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/j3/prod /var/backups/j3
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
```

`/etc/j3/prod.env`:

```bash
J3_OWNER_EMAIL=admin@example.com
J3_OWNER_NAME=J3 Owner
# Set once, then remove from disk after first login and rotate in the app.
J3_OWNER_PASSWORD=replace-with-a-long-random-secret
J3_CORS_ORIGIN=https://app.example.com,capacitor://localhost,http://localhost
```

Staging uses `staging.example.com`, `staging-api.example.com`, port `8788` and its own database.
Never point staging at the production SQLite file.

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now j3d-prod j3d-staging
```

## Caddy HTTPS + SSE

`/etc/caddy/Caddyfile` — `flush_interval -1` is essential for live scoring SSE.

```caddyfile
app.example.com {
    root * /srv/j3/prod
    encode zstd gzip
    file_server
    try_files {path} /index.html
}

staging.example.com {
    root * /srv/j3/staging
    encode zstd gzip
    file_server
    try_files {path} /index.html
}

api.example.com {
    encode zstd gzip
    reverse_proxy /api/* 127.0.0.1:8787 { flush_interval -1 }
    reverse_proxy /uploads/* 127.0.0.1:8787
    header { Strict-Transport-Security "max-age=31536000; includeSubDomains" }
}

staging-api.example.com {
    encode zstd gzip
    reverse_proxy /api/* 127.0.0.1:8788 { flush_interval -1 }
    reverse_proxy /uploads/* 127.0.0.1:8788
}

downloads.example.com {
    root * /srv/j3/downloads
    file_server
}
```

Point DNS A/AAAA records at the VM. Caddy obtains and renews certificates automatically.

## Mobile clients against a server

Three local mobile flavours are available (see `docs/CLI.md`):

```bash
bun run build:android:dev                       # local venue server via scripts/mobile.sh url
bun run build:android:qa                        # standalone seeded APK, no server
J3_SERVER_URL=https://api.example.com bun run build:android:prod  # live/production
```

The **QA** build bundles the seeded SQLite database and read-only API snapshot, so it opens with
data and signs in as admin on first launch. The **dev** and **production** builds talk only to their
configured server; production requires `J3_SERVER_URL` and keeps normal participant-only phone scope.
**Offline write sync is not enabled yet** — scoring writes require the venue server; keep the server
backed up and reachable during events.

Hosted tagged releases use the production flavour with the public repository variable
`J3_SERVER_URL` baked in. Put the signed release APK (or a QA APK for sideloading) under
`/srv/j3/downloads/android/`. iOS QA is easiest via TestFlight; an ad-hoc `.ipa` needs registered
device UDIDs and a distribution profile, and a raw IPA link will not install like an APK.

For automatic versioned downloads, set the repository **variable** `J3_SERVER_URL` to the public API
origin (e.g. `https://api.example.com`). Every `v*` tag then publishes `j3-android-<version>.apk`,
`j3-ios-<version>.app.zip` and, when Apple secrets exist, `j3-ios-<version>.ipa` on the GitHub
release. The APK is release-signed; stable upgrades need the Android keystore secrets from
`docs/ENVIRONMENT.md`. Leave `J3_SERVER_URL` blank if the API is served from the same origin as the
web client.

## SQLite backups

```bash
#!/usr/bin/env bash
set -euo pipefail
ts=$(date -u +%Y%m%dT%H%M%SZ)
sqlite3 /var/lib/j3/prod/j3.db ".backup '/var/backups/j3/prod-$ts.db'"
sqlite3 /var/lib/j3/staging/j3.db ".backup '/var/backups/j3/staging-$ts.db'"
find /var/backups/j3 -name '*.db' -mtime +14 -delete
```

Run nightly with a systemd timer and copy encrypted backups to OCI Object Storage (20 GB free). Use
SQLite `.backup`/`VACUUM INTO`, never a raw copy while writes are active, and test restores.

## Caveats

- Oracle may reclaim Always Free compute that is idle for 7 days; real traffic avoids this.
- Home region is fixed at signup. Choose carefully.
- Single VM is a single point of failure; backups and monitoring are your responsibility.
- Restrict SSH to your IP, keep `J3_OWNER_PASSWORD` out of source control, and use HTTPS everywhere.
