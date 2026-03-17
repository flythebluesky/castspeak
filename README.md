# CastSpeak

Play text-to-speech on Google Nest Mini (and other Chromecast) devices via CLI or REST API. No API keys or cloud credentials required — uses Google Translate TTS and the Chromecast CASTv2 protocol over your local network.

## Install

```bash
go install .
```

Or build locally:

```bash
go build -o castspeak .
```

## CLI Usage

```bash
# List Cast devices on your network
castspeak devices

# Scan for devices without mDNS (useful when mDNS is blocked)
castspeak scan --timeout 15
castspeak scan --timeout 15 --save   # save results for offline use

# View/manage saved devices
castspeak devices saved
castspeak devices forget

# Speak on a device
castspeak speak --device "Living Room speaker" --text "Dinner is ready"

# Speak in another language
castspeak speak --device "Kitchen" --text "Bonjour le monde" --language fr

# Target by UUID instead of name
castspeak speak --uuid "abc-123" --text "Hello"

# Play an audio URL
castspeak play --device "Office" --url "https://example.com/chime.mp3"

# Volume control
castspeak volume --device "Bedroom" --level 0.3
castspeak mute --device "Bedroom"
castspeak unmute --device "Bedroom"

# Stop playback
castspeak stop --device "Living Room speaker"

# Check device status
castspeak status --device "Kitchen"

# Start the HTTP server
castspeak serve
castspeak serve --port 9090
```

Run `castspeak <command> --help` for all flags.

## REST API

Start the server with `castspeak serve`, then:

### `GET /devices?timeout=5`

```json
{
  "devices": [
    {
      "name": "Living Room speaker",
      "uuid": "abc-123",
      "addr": "192.168.1.42",
      "port": 8009,
      "model": "Google Nest Mini"
    }
  ]
}
```

### `POST /speak`

```bash
curl -X POST http://localhost:8080/speak \
  -H "Content-Type: application/json" \
  -d '{"text": "Hello from the API", "device_name": "Living Room speaker"}'
```

```json
{"status": "ok", "device": "Living Room speaker", "chunks": 1}
```

**Body fields:**
- `text` (required) — up to 5000 characters
- `device_name` or `device_uuid` (at least one required)
- `language` — defaults to `"en"`

## How It Works

1. Discovers Cast devices on the local network via mDNS, falling back to saved devices when mDNS is blocked (e.g. by CrowdStrike Falcon). Use `castspeak scan --save` to populate saved devices via TCP port scan.
2. Builds Google Translate TTS URLs, splitting long text into ~200-character chunks at sentence boundaries
3. Connects to the target device over the CASTv2 protocol (TLS on port 8009) and plays each chunk sequentially

## Dependencies

- [go-chromecast](https://github.com/vishen/go-chromecast) — Cast protocol and mDNS discovery
- [chi](https://github.com/go-chi/chi) — HTTP router (server mode only)
