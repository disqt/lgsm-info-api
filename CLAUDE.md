# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build — must use -o; `go build ./cmd/main.go` produces ./main (file name),
# not ./lgsm-info-api (module name), which the systemd unit won't pick up.
go build -o lgsm-info-api ./cmd/main.go

# Run locally (listens on :8080)
go run ./cmd/main.go

# Run tests
go test ./...

# Run single test
go test ./cmd -run TestGetServersHandler
```

## Deployment

The API runs as a systemd service on the production VPS at `/home/dev/projects/lgsm-info-api/`.

```bash
# Check status
systemctl status lgsm-info-api.service

# Restart after changes
sudo systemctl restart lgsm-info-api.service
```

After restart, verify the new binary is actually running — `is-active` is not enough:
```bash
stat -c '%y' /home/dev/projects/lgsm-info-api/lgsm-info-api  # mtime should be recent
sudo journalctl -u lgsm-info-api.service -n 5 --no-pager     # fresh startup logs
```

## Architecture

This is a Go/Gin HTTP API that queries game server status using the external `gamedig` CLI tool, plus two non-gamedig sources: a direct file read for Windrose and an HTTP call to Palworld's own REST API. Neither game is queryable through gamedig.

**Request flow:**
1. `GET /servers` hits `GameServersHandler` in `cmd/main.go`
2. Handler calls `gameServers.GetGameServers()` which iterates over hardcoded `serverLookups`
3. For each server, `GameDigClient` executes `gamedig --type <game> <host> [--port <port>]`
4. Separately, `GetWindroseServer()` reads the local WindrosePlus `server_status.json` file, and `GetPalworldServer()` calls the Palworld REST API on localhost
5. Responses are parsed and transformed into `OnlineGameServer` or `OfflineGameServer`
6. `model.NewResponse()` in `cmd/model/response.go` builds the final JSON response

**Key files:**
- `pkg/gameServers/gameServerService.go` - Server lookup definitions and query orchestration
- `pkg/gameServers/model/gameServer.go` - Game server domain models with Steam/redirect URL generation
- `pkg/gameServers/client/gameDigClient.go` - GameDig CLI wrapper (injectable for testing)
- `pkg/gameServers/client/windroseClient.go` - WindrosePlus status-file reader (injectable for testing)
- `pkg/gameServers/client/palworldClient.go` - Palworld REST API client (injectable for testing)
- `cmd/model/response.go` - API response transformation

**External dependency:** Requires `gamedig` CLI installed on the system (npm package `gamedig`). For Windrose to appear online, the WindrosePlus status file at `/home/windrose/windrose/server-files/windrose_plus_data/server_status.json` must be readable by the API process and updated within 90s (matches the freshness gate used by `windrose-metrics.sh`).

For Palworld to appear online, `PALWORLD_ADMIN_PASSWORD` must be set in the systemd `EnvironmentFile` (`/etc/lgsm-info-api.env`) and match `AdminPassword` in the server's `PalWorldSettings.ini`. `PALWORLD_ADMIN_USER` is optional and defaults to `admin`. With no password set the API logs a warning once at startup and reports Palworld offline forever rather than issuing requests that can only 401.

## Server Configuration

Server definitions are hardcoded in `pkg/gameServers/gameServerService.go` as `serverLookups` (gamedig-queried), plus the Windrose (file-read) and Palworld (REST) constants. Each gamedig server has game ID, host, and optional port.

Steam connect URLs for CS2 use the `steam://rungameid/730//+connect` format to work around Steam's hostname DNS resolution bug with the standard `steam://connect/` protocol.

Palworld is not gamedig-queryable either, but for a different reason: the dedicated server only starts its Steam A2S responder when launched with `-publiclobby`, and this one deliberately omits that flag (private server, and it keeps Palworld off port 27015, which is CS2's). The query port is bound but silent, so `gamedig --type palworld` fails even from localhost. Status therefore comes from Palworld's REST API on `127.0.0.1:8212`: `/v1/api/metrics` for player counts and `/v1/api/info` for the server name. A failing `/info` does not mark the server offline, since a metrics response already proves it is up. Unlike Windrose, Palworld has a joinable address to display (`disqt.com:8211`), and its redirect points at the `disqt.com/palworld` landing page because Palworld has no working `steam://` connect scheme.

Windrose has no equivalent connect link (Unreal-based, not Steam Source) and no usable A2S responder, so it's queried by reading the WindrosePlus dashboard's local `server_status.json` file. The path and freshness gate are constants in `cmd/main.go`. The Windrose response intentionally has empty `Url` and `Redirect` — there's no joinable URL to copy and no one-click join scheme; players join via the in-game invite-code flow.

## Production Notes

- **nginx caching:** The `/servers` endpoint is cached by nginx with a 10-minute TTL and `stale-while-revalidate`, so clients may receive slightly stale data while a fresh response is being fetched in the background. Cache files live at `/var/cache/nginx/`. Force-purge after deploy: `sudo find /var/cache/nginx -type f -delete && sudo systemctl reload nginx`.
- **Concurrent queries:** `GetGameServers` fans out gamedig queries via goroutines + `sync.WaitGroup` (see `pkg/gameServers/gameServerService.go`). The Windrose file read in `cache.refresh()` happens sequentially after the WaitGroup joins — minor and not worth fanning out (local 500-byte file).
- **Server lookup order:** gamedig servers are queried in this order: minecraft, valheim, xonotic, csgo (CS2); Windrose and Palworld are appended after the WaitGroup joins. The final response is sorted by `Running` (online first) then alphabetically, so this order doesn't affect display ordering.
