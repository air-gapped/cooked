## Context

cooked fetches upstream URLs over HTTPS. In air-gapped Kubernetes clusters, upstreams use TLS certificates signed by internal CAs. These CA certificates are distributed to all namespaces by trust-manager (cert-manager project) as ConfigMap volume mounts — they are not known at image build time.

Currently the Dockerfile runs `update-ca-certificates` at build time, baking only public CAs into the image. Go's `crypto/x509` loads the system cert pool once via `sync.Once` and caches it for the process lifetime, so certs must exist on disk before the first outbound TLS connection.

## Goals / Non-Goals

**Goals:**
- Runtime-mounted CA certificates are trusted without rebuilding the image
- Zero-config for deployments without custom CAs (no-op behavior)
- Document the Kubernetes volume mount pattern for trust-manager users

**Non-Goals:**
- Hot-reloading of CA certificates after the Go process starts (Go does not support this)
- Supporting non-PEM certificate formats (JKS, PKCS12) — Go only needs PEM
- Replacing the existing build-time CA injection pattern (both coexist)

## Decisions

### 1. Shell entrypoint script over init container

Run `update-ca-certificates` in the main container entrypoint before `exec cooked`, rather than requiring users to configure an init container.

**Rationale:** An entrypoint script is self-contained — it ships with the image and works without any Kubernetes-specific configuration. Init containers require users to understand the cert merge workflow and maintain extra pod spec. The entrypoint approach means mounting certs to `/usr/local/share/ca-certificates/` is all that's needed.

**Alternative considered:** Documenting an init container pattern. Rejected because it pushes complexity to the user and is easy to misconfigure.

### 2. Suppress stderr from `update-ca-certificates`

The script runs `update-ca-certificates 2>/dev/null || true`.

**Rationale:** On a clean startup with no custom certs, `update-ca-certificates` may emit warnings about already-processed certs. The `|| true` ensures a missing or broken `update-ca-certificates` binary doesn't prevent cooked from starting (defense in depth for custom base images).

### 3. `exec` to replace the shell process

The script uses `exec cooked "$@"` so that cooked becomes PID 1 and receives signals directly.

**Rationale:** Without `exec`, the shell stays as PID 1 and cooked runs as a child. Kubernetes sends SIGTERM to PID 1 — if that's the shell, cooked never receives the graceful shutdown signal.

## Risks / Trade-offs

- **~50ms startup overhead** from running `update-ca-certificates` on every container start → Acceptable for a long-running server process. No mitigation needed.
- **`readOnlyRootFilesystem: true` breaks `update-ca-certificates`** → Documented in README. Users must add an `emptyDir` on `/etc/ssl/certs/` if using read-only root.
- **`.crt` extension requirement is easy to miss** → Documented prominently. trust-manager key names are user-defined, so users must ensure the `subPath` target ends in `.crt`.
