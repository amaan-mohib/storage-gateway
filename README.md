# Storage Gateway

Lightweight storage gateway in Go with a gateway + worker architecture. Handles incoming uploads to any S3 compatible backend (used MinIO here), stores files in multiple backends (S3, Firebase), and runs background processing tasks (image/video optimization, backups) via a worker queue.

Key components

- `gateway/` — HTTP server, routing, and services.
- `worker/` — background worker that processes queue tasks (uploads, backups, optimizations).
- `storage/` — storage adapters (S3, Firebase) and optimizer helpers (image/video).
- `queue/` — enqueue + producer + task definitions.

Requirements

- Go 1.20+ (or compatible toolchain)
- `ffmpeg` on PATH (used by video/image optimization)
- `libvips` installed on the host (required for image optimization).
- Docker & docker-compose (optional, for containerized runs)
- AWS credentials (for S3) and Firebase service account (for Firebase Storage)

Quickstart (development)

1. Ensure dependencies and tools are installed:

```bash
go version
which ffmpeg
```

2. Build and run the gateway locally:

```bash
cd gateway
go run ./src
```

3. Build and run the worker (separate process):

```bash
cd gateway
go build -o worker_bin ./worker
./worker_bin
```

Using Docker / docker-compose

From the repo root you can start services with docker-compose (builds images if needed):

```bash
docker-compose up --build
```

Configuration & secrets

- App configuration lives under `gateway/src/config`.
- Secrets layout: place provider credentials under a secrets root path and a logical bucket name. The repository expects secrets in the format:

  `$SECRETS_PATH/<bucket>/<firebase.json|s3_credentials>`

  Examples:
  - `$SECRETS_PATH/primary-bucket/firebase.json` — Firebase service account JSON for the `primary-bucket` backend.
  - `$SECRETS_PATH/primary-bucket/s3_credentials` — S3 credentials file (format your deployment expects) for the same bucket.

- Firebase credentials: a service account JSON file (see example path above).
- S3 / AWS credentials: should be placed in the secrets path as an `s3_credentials` file used by your deployment.

Storage backends

- Primary S3 store: see `gateway/storage/s3_store`.
- Firebase store: see `gateway/storage/firebase_store` (requires service account JSON).
- Primary MinIO/S3 fallback: when a file is not found in the primary MinIO store, the application will try to fetch it from the first available backup store (e.g., secondary S3 or Firebase backup) according to the configured backup order.

Worker & queue

- The worker consumes tasks defined in `gateway/queue/tasks.go` and processing logic in `gateway/worker/handler`.
- Queue producer and enqueue helpers are in `gateway/queue`.

Development notes

- If you change code in `gateway/`, re-build the gateway and/or worker binaries.
- Ensure `ffmpeg` and `libvips` are available for video/image optimization steps (used inside `gateway/storage/optimizer`).

Contributing

- Open an issue or PR with a focused change; keep commits small and tests passing (though no tests yet).
