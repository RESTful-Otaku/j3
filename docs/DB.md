# Database clean / purge / reseed / rebuild

`scripts/db.sh`, `J3_DB`/`J3_DB_PATH` (default `./data/j3-demo.db`), removes `-wal`/`-shm`.
```
bun run db:fresh | db:purge | db:reseed | db:seed | db:reset | db:rebuild | db:full | db:existing | db:backup
```
`j3d reset` removes db+wal+shm then reseeds (`-okami`, `-full`); `j3d seed`. `bash scripts/db.sh rebuild` fresh full. `db:existing` no-op. Backup `db:backup`. Single SQLite file portable. QA APK regenerates `j3-demo.db` + `j3-demo-data.json`.
