# fxcontrol [![Build and release](https://github.com/danesparza/fxcontrol/actions/workflows/release.yaml/badge.svg)](https://github.com/danesparza/fxcontrol/actions/workflows/release.yaml) 

Discovery API and Controller UI for FX services on the local network.

## Layout

The layout follows fxaudio and fxpixel: `main.go` is the entry point, `cmd/`
contains CLI commands, `api/` contains HTTP handlers and routing, `internal/`
contains implementation packages, `docs/` contains generated Swagger docs,
`version/` contains build version metadata, and `dist/` contains packaging assets.
Discovery lives in `internal/discovery`, advertisement conversion in
`internal/mapper`, response models in `internal/model`, and the HTTP lifecycle
in `internal/server`.

## Run

```sh
go run . start --listen :3090
curl http://localhost:3090/v1/discover/
```

`GET /v1/discover/` returns the latest in-memory snapshot
immediately. Discovery scans `_fx._tcp.local.` at startup and every 30 seconds,
with a five-second scan window. It recognizes the shared version 1 advertisements
from fxaudio, fxpixel, fxdmx, and fxtrigger. Multicast DNS must be available on the
local network.

Each successful scan replaces the previous snapshot, removing services no longer
found. Failed scans retain the last successful snapshot and log a warning. Before
the first scan completes, or when no services are found, `data` is `[]`.
Duplicate advertisements are combined by service type and ID; responses are
sorted by those keys. The cache is not persisted across restarts.

Example response:

```json
{
  "message": "Discovered FX services",
  "data": [
    {
      "id": "fxaudio:stage:3030",
      "name": "fxaudio-stage-3030",
      "service": "fxaudio",
      "host": "stage.local.",
      "port": 3030,
      "addresses": ["192.168.1.10"],
      "api": "v1",
      "scheme": "http",
      "path": "/v1"
    }
  ]
}
```

Network addresses and metadata are returned as advertised; discovery does not
probe API health. IPv6 link-local addresses may require an interface scope when
used by a client. SIGINT/SIGTERM stop discovery and gracefully drain HTTP requests.

## Development

Requires Go 1.26.6 or newer. Mock generation additionally requires Docker.

```sh
make gen-mocks  # Run pinned mockery in a Go Docker image
make swagger   # Regenerate API documentation
go test ./...
make lint-new  # go vet and Go formatting checks
```
