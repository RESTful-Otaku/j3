# Environment, secrets and white-label configuration

`.env.example` is the committed, safe template for local QA and hosted production. Copy it to
`.env.local` (git-ignored) and load it with `set -a; source .env.local; set +a`. Never commit real
passwords, SSH keys, keystores or signing certificates.

## QA defaults

`bun run dev` uses the managed TUI with seeded Okami 2027 data, local SQLite, demo login and the
offline APK workflow. The template's `J3_DEMO=1`, `J3_SEED_DEMO=1`, `VITE_OFFLINE_DEMO=1` and
`VITE_QA_AUTOLOGIN=1` values describe that mode. They are safe placeholders, not credentials.

## Production overrides

| Setting | QA | Production | Where |
|---|---|---|---|
| `J3_ADDR` | `127.0.0.1:8787` | `0.0.0.0:8787` behind Caddy | server env / systemd |
| `J3_DB` | `./data/j3.db` | `/srv/j3/production/j3.db` | server env / systemd |
| `J3_SEED_DEMO` | `1` | `0` | server env |
| `J3_DEMO` | `1` | `0` | server env |
| `J3_OWNER_EMAIL` / `J3_OWNER_NAME` | `j3@local` / `J3 Owner` | organiser identity | server env |
| `J3_OWNER_PASSWORD` | blank (one-time generated) | strong secret | GitHub secret / systemd |
| `J3_CORS_ORIGIN` | localhost | public HTTPS origin | server env |
| `J3_SERVER_URL` / `VITE_API_BASE_URL` | blank | public HTTPS API URL | GitHub variable `J3_SERVER_URL` |
| `VITE_OFFLINE_DEMO` / `VITE_QA_AUTOLOGIN` | `1` | unset / `0` | build env (never set for release web builds) |

Set `J3_SERVER_URL` as a **repository variable**, not a secret: it is a public URL baked into the
released mobile clients. Leave it blank when the web client and API are served from the same HTTPS
origin.

## GitHub Actions

- **Actions variables**: `J3_SERVER_URL` — hosted API URL for versioned release APK/`.app`.
- **Deployment secrets**: `DEPLOY_HOST`, `DEPLOY_USER`, `DEPLOY_PATH`, `DEPLOY_SSH_KEY`, plus
  environment-specific `DEPLOY_HOST_DEV`, `DEPLOY_HOST_STAGING`, `DEPLOY_HOST_PRODUCTION` and
  `DEPLOY_PATH_*` when using the environment-gated workflow.
- **Android signing (optional)**: `ANDROID_KEYSTORE_BASE64`, `ANDROID_KEYSTORE_PASSWORD`,
  `ANDROID_KEY_ALIAS`, `ANDROID_KEY_PASSWORD`. Without them CI uses the runner's standard debug
  keystore so the release APK is still signed and installable, but upgrades need the stable key.
- **iOS signing (optional)**: `APPLE_CERTIFICATE_P12`, `APPLE_CERTIFICATE_PASSWORD`,
  `APPLE_PROVISIONING_PROFILE`, `APPLE_TEAM_ID`, with optional
  `APPLE_PROVISIONING_PROFILE_NAME` and `APPLE_EXPORT_METHOD`. Without them releases still include
  the unsigned simulator `.app`; signed `.ipa` for real devices is produced when configured.

Workflows use least-privilege `permissions` blocks, upload curated `j3-*` assets only, and never
print secret values. Apple/Android signing material stays in GitHub encrypted secrets.

## White-label identity

The setup wizard stores the event `title`, `logoPath`, `bannerPath` and theme tokens. The client
loads them from `/api/v1/demo/event` (or the bundled offline snapshot) and applies them as CSS
variables, so changing the organiser's title, banner, logo, main colour or secondary colour does not
require a code change. The J3 licence notice is always retained.
