# Agent workflow (keep agents on track)

1. Read `AGENTS.md` §§1–3, §11–14. Keep it current; update §11–14.
2. **Principles**: everything data-driven; offline-first; deterministic auditable; config-not-constants; accessible; small slices; production quality; non-technical; automate; no bad states.
3. **Outstanding tracking**: before work, list `git status`, `git log`, current known gaps in §14. After, update change log + current state.
4. **Build flavours**: dev/qa/prod web + mobile; `J3_SERVER_URL` prod; `VITE_OFFLINE_DEMO` QA; `VITE_QA_FULL_PHONE`.
5. **DB**: `scripts/db.sh`. **Stress**: `scripts/stress.sh`. **Component QA**: `scripts/component-qa.sh`. **Gauntlet**: run all checks + CI/CD.
6. **No shortcuts**: make tests pass by fixing work; no hidden silencing; `go test -race`; `svelte-check` 0/0; generated types parity; `check:workflows`.
7. **Safe**: ConfirmDialog for permanent changes; preview; `notifySuccess(...,undo)`; `runUndo`; safe areas; no harmful; no secrets in cache/outbox; don't touch `.env`/secrets.
8. **Docs sync**: `bun run sync:api-docs`; `docs/*.md` -> `web/static/docs`.
9. **Don't** enable billing/branch protection; workflows `if visibility=='public'`; `check:workflows`.
10. **Mobile** independent: phone scope; drawer; safe areas; offline; pagination; no cramped fight cards; reliable sync.
