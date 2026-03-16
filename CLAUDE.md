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
main.go              # CLI entrypoint — subcommand dispatch (serve, speak, devices, volume, etc.)
internal/
  tts/               # Pure logic: Google Translate TTS URL builder + text chunking
  discovery/         # Wraps go-chromecast/dns for mDNS device discovery
  cast/              # Wraps go-chromecast/application for CASTv2 connection + playback
  speak/             # Shared orchestration layer — discovery + TTS + cast, used by both CLI and HTTP
  server/            # Chi HTTP router, handlers, JSON models
  cli/               # CLI flag parsing + subcommand runners
```

**Data flow:** `cli` or `server` → `speak` (orchestration) → `tts` + `discovery` + `cast`

## Key Conventions

- **No device cache** — fresh mDNS discovery per request. Avoids stale data from DHCP/power changes.
- **One cast connection per operation** — `withApp()` in `cast/cast.go` handles connect + defer close for all commands.
- **`speak` package is the shared layer** — all business logic goes through `speak.*`, never call `discovery`/`cast`/`tts` directly from handlers or CLI.
- All CLI subcommands use `flag.FlagSet` for consistent flag parsing.
- Device targeting uses `--device` (name) or `--uuid` — at least one required.

## Dependencies

- `github.com/vishen/go-chromecast` v0.3.4 — Cast protocol, mDNS. Not actively maintained; `app.Close()` is unsafe to call if `Start()` failed (nil TLS conn panic).
- `github.com/go-chi/chi/v5` — HTTP router, used only in `server/`.

## Testing

- `tts/` has 100% coverage — pure logic, most important to test thoroughly.
- `server/` tested via `httptest` — handler routing, JSON parsing, error cases.
- `discovery/`, `speak/` — validation and helper logic tested; network-dependent paths require real Cast devices.
- `cast/` — connection tests hit real network; only type/edge-case tests run offline.
