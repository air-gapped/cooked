## Why

cooked is designed for air-gapped environments where internal CAs sign upstream TLS certificates. The current Dockerfile runs `update-ca-certificates` at build time, but in Kubernetes deployments custom CA certs are distributed at runtime by trust-manager (cert-manager project) as ConfigMap volume mounts. Go's `crypto/x509` caches the system cert pool on first use (`sync.Once`), so certs must be present before the first TLS connection — a startup-time entrypoint script solves this.

## What Changes

- Add `entrypoint.sh` wrapper script that runs `update-ca-certificates` then `exec cooked`
- Update Dockerfile `ENTRYPOINT` to use the wrapper script instead of invoking `cooked` directly
- Document the Kubernetes runtime CA injection pattern in README.md (volume mount to `/usr/local/share/ca-certificates/`)

## Capabilities

### New Capabilities

- `runtime-ca-injection`: Container startup runs `update-ca-certificates` before the Go process, enabling runtime-mounted CA certificates (e.g. from trust-manager) to be trusted without rebuilding the image.

### Modified Capabilities

<!-- No existing specs to modify -->

## Impact

- **Dockerfile**: Entrypoint changes from direct binary to shell script — adds ~50ms startup overhead for `update-ca-certificates`
- **Container image**: Now includes `entrypoint.sh` in `/usr/local/bin/`
- **Backwards compatibility**: Existing deployments without custom CAs are unaffected — `update-ca-certificates` is a no-op when no new certs are mounted
- **Security posture**: No change to default TLS verification behavior; this provides a supported path to avoid `--tls-skip-verify`
