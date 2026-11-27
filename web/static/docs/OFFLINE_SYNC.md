# Offline & synchronisation

J3 is offline-first. Reads fall back; writes are queued; reconnect replays; server is authoritative.

## Read reliability
- `GET` success caches `j3-read-cache` (last-known-good). On network failure, **5xx**, 408, 429, non-JSON, Capacitor HTML -> serve cache. **Not** on 401/403/404 or `password_change_required`. No secrets cached.
- QA offline builds (`VITE_OFFLINE_DEMO=1`, Capacitor) bundle `web/static/j3-demo-data.json` + `j3-demo.db`; `send` serves bundled snapshot on non-OK/non-JSON; Capacitor 200 HTML handled.
- `livePoll`, `onLiveChange`, `state.changed`, `j3-session-changed`, replay `syncReplayUrl` re-pull on reconnect. `online` event flushes outbox.

## Write outbox
- Non-GET/HEAD on **network/Capacitor unreachable** are queued `j3-write-outbox` (`id,url,method,body,ts`). `{queued-offline:true}` returned so UI isn't blank. 
- `flushOutbox()` on `online` + after a successful response; replays `fetch` with fresh auth. `409` treated as already-applied. Failed entries retained. `outboxSize`, `clearOutbox`.
- **Not** queued on `401/403/404`/server `!ok` (server saw request); server remains authoritative; conflict resolution is server-side. QA `VITE_OFFLINE_DEMO` has richer `offlineWriteResponse` simulation + `j3-offline-state`.
- Spinners/placeholders; no silent blanks; `aria-live`.
