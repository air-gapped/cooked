## ADDED Requirements

### Requirement: Entrypoint runs update-ca-certificates before starting cooked

The container entrypoint SHALL run `update-ca-certificates` before executing the cooked binary. This ensures any CA certificate files mounted into `/usr/local/share/ca-certificates/` at runtime are added to the system trust store before Go's `crypto/x509` caches the cert pool.

#### Scenario: Custom CA cert mounted at runtime

- **WHEN** a PEM-encoded `.crt` file is mounted to `/usr/local/share/ca-certificates/` via a Kubernetes volume mount
- **THEN** `update-ca-certificates` incorporates it into `/etc/ssl/certs/ca-certificates.crt` before cooked starts, and cooked trusts upstreams signed by that CA

#### Scenario: No custom CA certs mounted

- **WHEN** no additional certificate files are present in `/usr/local/share/ca-certificates/` beyond the base image defaults
- **THEN** `update-ca-certificates` completes as a no-op and cooked starts normally with the default system CA bundle

#### Scenario: update-ca-certificates fails

- **WHEN** `update-ca-certificates` exits with a non-zero status (e.g. missing binary, permission error)
- **THEN** cooked SHALL still start with whatever certificates were available at build time

### Requirement: Entrypoint preserves signal handling

The entrypoint script SHALL use `exec` to replace itself with the cooked process, so that cooked runs as PID 1 and receives container signals (SIGTERM, SIGINT) directly.

#### Scenario: Graceful shutdown signal

- **WHEN** Kubernetes sends SIGTERM to the container
- **THEN** cooked receives the signal directly (not intercepted by a parent shell) and performs graceful shutdown

### Requirement: Entrypoint passes through all arguments

The entrypoint script SHALL forward all arguments to the cooked binary unchanged.

#### Scenario: Custom flags passed via container command

- **WHEN** the container spec includes additional arguments (e.g. `--allowed-upstreams=*.internal`)
- **THEN** those arguments are passed to cooked exactly as specified
