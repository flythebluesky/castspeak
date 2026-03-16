---
name: castspeak
description: Control Google Nest Mini and other Chromecast devices — speak text aloud, play audio URLs, set volume, mute/unmute, and check device status via the castspeak CLI. Use this skill whenever the user wants to interact with Cast devices, announce something on a speaker, play sounds on a Nest/Chromecast, or control smart speaker volume — even if they don't mention "castspeak" by name.
---

# CastSpeak Skill

Control Cast devices on the local network using the `castspeak` CLI.

## Prerequisites

The `castspeak` binary must be built and on your PATH (or run from the repo directory):

```bash
cd /Users/rori/_repos/castspeak && go build -o castspeak .
```

## Workflow

Always discover devices before issuing commands — device names come from mDNS and you need the exact name:

```bash
castspeak devices --timeout 5
```

Then use the device name from the output in subsequent commands.

## Commands

### Speak text aloud

```bash
castspeak speak --device "<device name>" --text "<message>" --language en
```

### Play an audio URL

```bash
castspeak play --device "<device name>" --url "<media url>"
```

### Volume control

```bash
castspeak volume --device "<device name>" --level 0.5
castspeak mute --device "<device name>"
castspeak unmute --device "<device name>"
```

`--level` is a float: 0.0 (silent) to 1.0 (max). So 30% = `--level 0.3`.

### Stop playback

```bash
castspeak stop --device "<device name>"
```

### Check device status

```bash
castspeak status --device "<device name>"
```

### Start the HTTP API server

```bash
castspeak serve --port 8080
```

Starts a REST API on the given port (default 8080, or `PORT` env var). Useful for integrations that prefer HTTP over CLI.

## Targeting devices

- `--device "<name>"` — match by device name (from `castspeak devices` output)
- `--uuid "<uuid>"` — match by device UUID (also shown in discovery output)
- At least one is required for all commands except `devices` and `serve`.

## Tips

- **Discover first.** Always run `castspeak devices` before other commands so you have the exact device name.
- **Keep text short.** Text is split into ~200-char chunks — shorter messages play faster. Max 5000 characters.
- **Language codes.** `--language` accepts BCP-47 codes: `en`, `fr`, `es`, `de`, `ja`, etc. Default is `en`.
- **Timeout.** Discovery defaults to 5 seconds. If a device isn't found, try `--timeout 10`.
- **Each command rediscovers.** There's no persistent connection — each invocation does fresh mDNS discovery.
