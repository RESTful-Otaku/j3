# Versioning

**SemVer `x.y.z`.** `package.json`/`web/package.json` `version` currently `0.8.0`. `scripts/version.sh` strict-checks. 

## Channels
- `alpha`: `x.y.za` + strict SemVer `x.y.z-alpha`, build number `+0`.
- `beta`: `x.y.zb` + strict `x.y.z-beta`, `+1`.
- `release`: `x.y.z`, `+2`.
`J3_VERSION`, `J3_VERSION_NAME`, `J3_VERSION_CODE`, `J3_VERSION_CHANNEL`.
`scripts/version.sh [alpha|beta|release] [check|name|strict|code|core|print]`. 

## Mobile
Android `J3_VERSION_NAME`, `J3_VERSION_CODE`, `J3_VERSION_CHANNEL`. `mobile.sh`. Defaults `0.8.0`, `8002`. iOS `MARKETING_VERSION`, `CURRENT_PROJECT_VERSION`. Hosted `mobile.yml`, `release.yml` `v*`; `J3_VERSION_NAME`, `J3_VERSION_CODE`. 
## CLI/server
`j3d -version`, `J3_VERSION`, `config`. `release.sh`, `local-build`. 
## Licence/copyright
`LICENSE` PolyForm Noncommercial 1.0.0. `Copyright © 2026 Jonathan Harrison (RESTful_Otaku)`. Required notice. package `license`,`author`. Not sold. 
