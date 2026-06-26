<p align="center">
<img height="44" alt="image" src="https://github.com/user-attachments/assets/75c10d10-5ebc-4f6a-b762-d0e971af7f97" />
</p>

# importarr

Force import stuck queue items in Sonarr and Radarr that fail automatic import.
This tool target specifically items that have this message in sonarr/radarr

> "Found matching series via grab history, but release was matched to series by ID. Automatic import is not possible. See the FAQ for details."

According to Sonarr and Radarr documentation

> This error in Sonarr or Radarr occurs when a downloaded release is matched using an external ID (like a TVDB or TMDb ID), but the actual release name fails to perfectly match the expected series title.

these items matches 99% of the time and mostly are due to obfuscated names, messed up release name or TVDB and TBDb did not have the alias registered in their databas (Usually anime)

## Features
- support for multiple Arr instances via Yaml file.
- Fallback feature where failed imports are blocklisted and removed. 
- TUI interface for manual import .
- Tell me !!

## Installation

```bash
go install github.com/ajmandourah/importarr@latest
```

Or build from source:

```bash
go build -o importarr .
```

## Configuration

### .env file (default)

Copy `.env.example` to `.env` and fill in your credentials:

```env
SONARR_URL=http://localhost:8989
SONARR_API_KEY=your_api_key_here

RADARR_URL=http://localhost:7878
RADARR_API_KEY=your_api_key_here
```

### config.yaml (optional)

For multiple instances, create `config.yaml` in the working directory:

```yaml
instances:
  - name: sonarr-main
    type: sonarr
    url: http://localhost:8989
    api_key: your_api_key_here

  - name: radarr-main
    type: radarr
    url: http://localhost:7878
    api_key: your_api_key_here
```

## Usage

### Scan for stuck items

```bash
importarr scan
importarr scan -n sonarr-main    # target specific instance
importarr scan -a                # target all instances
importarr scan -i                # interactive TUI mode
```

### Force import

```bash
importarr import
importarr import -n sonarr-main  # target specific instance
importarr import -a              # target all instances
importarr import -f              # remove + re-search on failure
importarr import -i              # interactive TUI mode
```

### Continuous watch mode

```bash
importarr watch                  # default: every 10 minutes
importarr watch -t 5m            # every 5 minutes
importarr watch -t 1h -f         # every hour with fallback
```

### List configured instances

```bash
importarr config list
```

## Docker

Run in watch mode:

```bash
docker run -v $(pwd)/.env:/app/.env ghcr.io/ajmandourah/importarr:latest watch -t 5m
```

Build locally:

```bash
docker build -t importarr .
docker run -v $(pwd)/.env:/app/.env importarr watch -t 5m
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | Detect stuck queue items |
| `import` | Force import stuck items |
| `watch` | Continuously scan and import at interval |
| `config list` | List configured instances |

## Flags

| Flag | Description |
|------|-------------|
| `-n, --instance` | Target specific instance by name |
| `-a, --all` | Target all configured instances |
| `-i, --interactive` | Launch interactive TUI |
| `-f, --fallback` | Remove from queue and trigger search on import failure |
| `-t, --interval` | Watch mode scan interval (default: 10m) |

## How It Works

1. Fetches queue items matching the "via grab history" error
2. Retrieves available import files via `/api/v3/manualimport`
3. Sends a `ManualImport` command to force the import
4. Optionally removes from queue and triggers search if import fails

## License

MIT
