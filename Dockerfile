FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git=2.52.0-r0 curl=8.17.0-r1

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# Download embedded assets with SHA-256 integrity verification (F-08)
COPY Makefile ./
COPY embed/ embed/
SHELL ["/bin/ash", "-eo", "pipefail", "-c"]
RUN mkdir -p embed && \
    curl -fsSL -o embed/mermaid.min.js "https://cdn.jsdelivr.net/npm/mermaid@11.12.2/dist/mermaid.min.js" && \
    curl -fsSL -o embed/github-markdown-light.css "https://raw.githubusercontent.com/sindresorhus/github-markdown-css/v5.9.0/github-markdown-light.css" && \
    curl -fsSL -o embed/github-markdown-dark.css "https://raw.githubusercontent.com/sindresorhus/github-markdown-css/v5.9.0/github-markdown-dark.css" && \
    echo "d0830a6c05546e9edb8fe20a8f545f3e0dc7c4c3134d584bad9c13a99d7a71e0  embed/mermaid.min.js" | sha256sum -c && \
    echo "de2d14b5290b8cf2af74c95e92560d9c00642ae72de0b856cece3e4eddb2d885  embed/github-markdown-light.css" | sha256sum -c && \
    echo "b45ead2db01f5856c4eb378f21f47da63f6b0ecf3be5d06385472164b7283df6  embed/github-markdown-dark.css" | sha256sum -c

COPY . .
RUN cp README.md embed/project-readme.md

ARG LDFLAGS=""
RUN CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o /cooked ./cmd/cooked

FROM alpine:3.23

RUN apk add --no-cache ca-certificates=20251003-r0 && update-ca-certificates

WORKDIR /app

COPY --from=builder /cooked /usr/local/bin/cooked

EXPOSE 8080

ENTRYPOINT ["cooked", "--listen", "0.0.0.0:8080"]
