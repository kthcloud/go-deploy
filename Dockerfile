# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM alpine:3.22 AS runner-setup

RUN apk add --no-cache ca-certificates \
  && addgroup -S go-deploy \
  && adduser  -S go-deploy \
  -G go-deploy \
  -h /opt/go-deploy \
  -s /sbin/nologin \
  -D \
  && mkdir -p /opt/go-deploy \
  && chown -R go-deploy:go-deploy /opt/go-deploy

FROM --platform=$BUILDPLATFORM golang:latest AS build-0

WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if needed
COPY go.mod go.sum ./

ENV CGO_ENABLED=0

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  go mod download -x

# Copy the rest of the project
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-1

ARG TARGETOS
ARG TARGETARCH

RUN --mount=type=cache,target=/go/pkg/mod \
  --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -ldflags="-w -s" -o go-deploy_${TARGETOS}_${TARGETARCH} ./cmd/go-deploy

# Runner
FROM scratch

COPY --from=runner-setup /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=runner-setup /etc/passwd /etc/passwd
COPY --from=runner-setup /etc/group /etc/group
COPY --from=runner-setup --chown=go-deploy:go-deploy /opt/go-deploy /opt/go-deploy

ENV HOME=/opt/go-deploy

WORKDIR /opt/go-deploy

ARG TARGETOS
ARG TARGETARCH

# Copy the binary and static assets from the builder stage
COPY --from=build-1 --chmod=755 /app/go-deploy_${TARGETOS}_${TARGETARCH} /usr/bin/go-deploy
COPY --from=build-1 --chown=go-deploy:go-deploy /app/index index

ARG RUN_USER=go-deploy
USER ${RUN_USER}

# Set environment variables and expose necessary port
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080

# Run the Go Gin binary
ENTRYPOINT ["/usr/bin/go-deploy"]
