##
## Provider stages:
##

# Provide ca-certificates
FROM alpine:latest AS ca-cert-provider

RUN apk add --no-cache ca-certificates

##
## Builder stages
##

# Download dependencies stage:
FROM --platform=$BUILDPLATFORM golang:latest AS builder-1

WORKDIR /app

COPY go.* ./

RUN --mount=type=cache,target=/go/pkg/mod/ \
    go mod download -x

# Copy source code stage:
FROM --platform=$BUILDPLATFORM builder-1 AS builder-2

COPY . .

# Compilation stage:
#   This is a separate stage since it asks for the TARGETOS and TARGETARCH
#   hence docker will run this stage for every TARGETOS and TARGETARCH
#   We want the previous stage (the download stage) to be separate so we
#   dont run it multiple times, since that would be unneccesary
FROM --platform=$BUILDPLATFORM builder-2 AS builder-3

ARG TARGETARCH
ARG TARGETOS

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s" \
    -trimpath \
    -o go-deploy_${TARGETOS}_${TARGETARCH} \
    .

##
## Runner stage
##

FROM scratch AS runner

ARG TARGETARCH
ARG TARGETOS

WORKDIR /opt/go-deploy

COPY --from=ca-cert-provider /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder-3 /app/index index
COPY --from=builder-3 /app/docs docs

COPY --from=builder-3 --chmod=755 /app/go-deploy_${TARGETOS}_${TARGETARCH} /usr/bin/go-deploy

ENV PORT=8080
ENV GIN_MODE=release

EXPOSE 8080

ENTRYPOINT [ "/usr/bin/go-deploy" ]
