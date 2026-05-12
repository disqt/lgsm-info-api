# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build
go build ./cmd/main.go

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

## Architecture

This is a Go/Gin HTTP API that queries game server status using the external `gamedig` CLI tool, plus a direct file read for Windrose (which gamedig does not support).

**Request flow:**
1. `GET /servers` hits `GameServersHandler` in `cmd/main.go`
2. Handler calls `gameServers.GetGameServers()` which iterates over hardcoded `serverLookups`
3. For each server, `GameDigClient` executes `gamedig --type <game> <host> [--port <port>]`
4. Separately, `GetWindroseServer()` reads the local WindrosePlus `server_status.json` file
5. Responses are parsed and transformed into `OnlineGameServer` or `OfflineGameServer`
6. `model.NewResponse()` in `cmd/model/response.go` builds the final JSON response

**Key files:**
- `pkg/gameServers/gameServerService.go` - Server lookup definitions and query orchestration
- `pkg/gameServers/model/gameServer.go` - Game server domain models with Steam/redirect URL generation
- `pkg/gameServers/client/gameDigClient.go` - GameDig CLI wrapper (injectable for testing)
- `pkg/gameServers/client/windroseClient.go` - WindrosePlus status-file reader (injectable for testing)
- `cmd/model/response.go` - API response transformation

**External dependency:** Requires `gamedig` CLI installed on the system (npm package `gamedig`). For Windrose to appear online, the WindrosePlus status file at `/home/windrose/windrose/server-files/windrose_plus_data/server_status.json` must be readable by the API process and updated within 90s (matches the freshness gate used by `windrose-metrics.sh`).

## Server Configuration

Server definitions are hardcoded in `pkg/gameServers/gameServerService.go` as `serverLookups` (gamedig-queried) and `windroseLookup` (file-read). Each gamedig server has game ID, host, and optional port.

Steam connect URLs for CS2 use the `steam://rungameid/730//+connect` format to work around Steam's hostname DNS resolution bug with the standard `steam://connect/` protocol.

Windrose has no equivalent connect link (Unreal-based, not Steam Source) and no usable A2S responder, so it's queried by reading the WindrosePlus dashboard's local `server_status.json` file. The path and freshness gate are constants in `cmd/main.go`. The Windrose response intentionally has empty `Url` and `Redirect` — there's no joinable URL to copy and no one-click join scheme; players join via the in-game invite-code flow.

## Production Notes

- **nginx caching:** The `/servers` endpoint is cached by nginx with a 10-minute TTL and `stale-while-revalidate`, so clients may receive slightly stale data while a fresh response is being fetched in the background.
- **Sequential queries:** The API queries each game server sequentially via the `gamedig` CLI (see the `for` loop in `GetGameServers`), so total response time scales linearly with the number of servers in `serverLookups`.
- **Server lookup order:** Servers are queried in this order: minecraft, valheim, xonotic, csgo (CS2), windrose. The valheim entry still exists in the lookup list, but the Valheim server is no longer running -- it will always appear as offline in the response. The final response is sorted by `Running` (online first) then alphabetically, so this order doesn't affect display ordering.
