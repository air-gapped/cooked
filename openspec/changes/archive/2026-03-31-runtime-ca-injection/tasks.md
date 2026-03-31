## 1. Entrypoint Script

- [x] 1.1 Create `entrypoint.sh` that runs `update-ca-certificates 2>/dev/null || true` then `exec cooked "$@"`
- [x] 1.2 Ensure script is executable (`chmod +x`)

## 2. Dockerfile

- [x] 2.1 Add `COPY entrypoint.sh /usr/local/bin/entrypoint.sh` to the runtime stage
- [x] 2.2 Change `ENTRYPOINT` from `["cooked", ...]` to `["entrypoint.sh", ...]`

## 3. Documentation

- [x] 3.1 Add "Runtime injection (Kubernetes)" subsection to README.md under Internal CA certificates
- [x] 3.2 Document the volume mount pattern with trust-manager ConfigMap example
- [x] 3.3 Document the three gotchas: `.crt` extension, `subPath` usage, `readOnlyRootFilesystem` emptyDir

## 4. Verification

- [x] 4.1 Build Docker image and verify cooked starts normally without custom certs
- [x] 4.2 Build Docker image, mount a test `.crt` file, verify `update-ca-certificates` picks it up
