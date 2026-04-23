# bazinga!

## Quickstart

```sh
# Pull & run the latest image
podman pull ghcr.io/gldraphael/bazinga:latest
podman run -it --rm -p 2222:2222 ghcr.io/gldraphael/bazinga:latest

# In another terminal, run
ssh -p 2222 user@localhost
```

## Other ways to run the app

```sh
# Run as CLI app
go run main.go

# Run as SSH Server
# Bash
BAZINGA__SSH__ENABLED=true go run main.go
# PowerShell
$env:BAZINGA__SSH__ENABLED="true"; go run main.go
```

## Config

The app can be configured using a `config.yaml` file and/or with environment variables:

```yaml
ssh:
  enabled: true            # BAZINGA__SSH__ENABLED:  If SSH must be enabled
  addr: "localhost:2222"   # BAZINGA__SSH__ADDR:     The address to listen for requests at
  host_key: "ssh_host_key" # BAZINGA__SSH__HOST_KEY: The ssh host key file
```

## Credits

Vibe coded with @basicbenjamin