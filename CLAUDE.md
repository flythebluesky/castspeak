# CastSpeak

Go CLI and REST API for playing text-to-speech on Google Nest Mini and other Chromecast devices via the CASTv2 protocol. No API keys needed.

## Build & Test

```bash
go build -o castspeak .
go test ./...
go vet ./...
```

## Architecture

```
main.go              # CLI entrypoint — subcommand dispatch (serve, speak, devices, scan, volume, etc.)
internal/
  tts/               # Pure logic: Google Translate TTS URL builder + text chunking
  discovery/         # mDNS device discovery + device persistence (store.go → ~/.config/castspeak/devices.json)
  scan/              # TCP port 8009 scan + HTTP 8008 eureka_info fetch — mDNS-free discovery
  cast/              # Wraps go-chromecast/application for CASTv2 connection + playback
  speak/             # Shared orchestration layer — discovery + scan + TTS + cast, used by both CLI and HTTP
  server/            # Chi HTTP router, handlers, JSON models
  cli/               # CLI flag parsing + subcommand runners
```

**Data flow:** `cli` or `server` → `speak` (orchestration) → `tts` + `discovery` + `scan` + `cast`

## Key Conventions

- **Fallback discovery chain** — `--host` (direct) → mDNS → saved devices. When mDNS is blocked (e.g. CrowdStrike Falcon), the system falls back to `~/.config/castspeak/devices.json` with a warning that addresses may be stale.
- **Network scan** — `castspeak scan` finds Cast devices via TCP port 8009 scan + HTTP 8008 `eureka_info`, bypassing mDNS entirely. Use `--save` to persist discovered devices.
- **One cast connection per operation** — `withApp()` in `cast/cast.go` handles connect + defer close for all commands.
- **`speak` package is the shared layer** — all business logic goes through `speak.*`, never call `discovery`/`cast`/`tts`/`scan` directly from handlers or CLI.
- All CLI subcommands use `flag.FlagSet` for consistent flag parsing.
- Device targeting uses `--device` (name) or `--uuid` — at least one required.

## Dependencies

- `github.com/vishen/go-chromecast` v0.3.4 — Cast protocol, mDNS. Not actively maintained; `app.Close()` is unsafe to call if `Start()` failed (nil TLS conn panic).
- `github.com/go-chi/chi/v5` — HTTP router, used only in `server/`.

## Testing

- `tts/` has 100% coverage — pure logic, most important to test thoroughly.
- `server/` tested via `httptest` — handler routing, JSON parsing, error cases.
- `discovery/` — validation, helper logic, and device persistence (store) fully tested; mDNS paths require real Cast devices.
- `scan/` — subnet enumeration, IP generation, and eureka_info parsing tested; actual port scanning requires real network.
- `speak/` — validation and helper logic tested; fallback chain and network-dependent paths require real Cast devices.
- `cast/` — connection tests hit real network; only type/edge-case tests run offline.
