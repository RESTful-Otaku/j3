# Local setup + env files (Linux / WSL)

J3 = Go `j3d` + pure SQLite + SvelteKit web + Capacitor Android/iOS. **Bun, never npm.** 

## Safe env templates (committed; copy & replace)
- `.env.example`, `.env.local.example` -> `.env.local`
- `.env.docker.example` -> `.env`
- `.env.staging.example` -> `.env.staging`
- `.env.production.example` -> `.env.production`
- `.env.mobile.example`, `.env.server.example`, `web/.env.example` -> `web/.env.local`
`.gitignore` ignores `.env`, `.env.*`, allows `.env.*.example`. **Never commit real secrets.** 

Load: `set -a; source .env.local; set +a`. Replace `CHANGE_ME`. 

## Modes
- **local dev**: `bun run dev` / `bun run db:*`; `VITE_OFFLINE_DEMO=0`, `J3_SEED_DEMO=1`, `J3_DEMO=1`.
- **QA**: `VITE_OFFLINE_DEMO=1`,`VITE_QA_AUTOLOGIN=1`,`VITE_QA_FULL_PHONE=1`, `scripts/mobile.sh`, `qa-docker.sh`. `build:qa`.
- **staging/beta**: `J3_VERSION_CHANNEL=beta`, `0.8.0b`, `J3_QA_API_URL`/`VITE_API_BASE_URL` staging, `J3_SEED_DEMO=0`,`J3_DEMO=0`, `VITE_*=0`.
- **production/live**: `J3_VERSION_CHANNEL=release`, `0.8.0`, `J3_SERVER_URL=https://api`, `VITE_API_BASE_URL`, `J3_SEED_DEMO=0`,`J3_DEMO=0`, `VITE_*=0`, `J3_PROD_API_URL`. 

## Backend bind/networking
`J3_ADDR` `127.0.0.1:8787` local, `0.0.0.0:8787` VPS/on-prem/cloud behind reverse proxy. `J3_DB`. `J3_CORS_ORIGIN` real HTTPS. `J3_SERVER_URL` public HTTPS API baked to clients. `VITE_API_BASE_URL`. `J3_DEV_API_URL`. Caddy/systemd. 
Real mobile devices on LAN need `http://<LAN-IP>:8787` (not `127.0.0.1`), CORS `http://localhost,capacitor://localhost,http://<LAN-IP>`, HTTPS in production. iOS real requires Xcode/macOS + signing; Linux/WSL Android only.

## Docker
`bash scripts/doctor.sh`. `bash scripts/qa-docker.sh full|okami|clean|light|dev`, `build`, `down`, `reset`, **`destroy`**, `status`, `logs`, `simulate`. `.env.docker`. auto runtime + doctor + port. **build up** + **complete teardown**. TUI `bun run dev` Docker QA. 

## `scripts/doctor.sh`
Checks git/bash/curl/python/Go/Bun/Java/adb/keytool/Docker/Podman/daemon. Tries `apt-get`, Bun install, `systemctl`/`service`, `podman machine`, prompts `Press Enter once done` and re-runs check. Most automated; direct user only for GUI/Docker Desktop/WSL/macOS Xcode. 

## `@`
- `@`. 
