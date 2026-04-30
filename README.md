# patchwork

Lightweight HTTP mock server that reads route definitions from a YAML config for local API development and testing.

## Installation

```bash
go install github.com/yourusername/patchwork@latest
```

## Usage

Define your routes in a `patchwork.yaml` file:

```yaml
routes:
  - path: /api/users
    method: GET
    status: 200
    body: |
      [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]

  - path: /api/users/:id
    method: GET
    status: 200
    body: |
      {"id": 1, "name": "Alice"}

  - path: /api/users
    method: POST
    status: 201
    body: |
      {"id": 3, "name": "Charlie"}
```

Start the mock server:

```bash
patchwork --config patchwork.yaml --port 8080
```

Your mock API is now running at `http://localhost:8080`.

```bash
curl http://localhost:8080/api/users
# [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `patchwork.yaml` | Path to the YAML config file |
| `--port` | `8080` | Port to listen on |

## License

MIT