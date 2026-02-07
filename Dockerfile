FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git curl

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

# Download embedded assets
COPY Makefile ./
COPY embed/ embed/
RUN mkdir -p embed && \
    curl -fsSL -o embed/mermaid.min.js "https://cdn.jsdelivr.net/npm/mermaid@11.12.2/dist/mermaid.min.js" && \
    curl -fsSL -o embed/github-markdown-light.css "https://raw.githubusercontent.com/sindresorhus/github-markdown-css/v5.9.0/github-markdown-light.css" && \
    curl -fsSL -o embed/github-markdown-dark.css "https://raw.githubusercontent.com/sindresorhus/github-markdown-css/v5.9.0/github-markdown-dark.css"

COPY . .

ARG LDFLAGS=""
RUN CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o /cooked ./cmd/cooked

FROM alpine:latest

RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /cooked /usr/local/bin/cooked

EXPOSE 8080

ENTRYPOINT ["cooked"]
