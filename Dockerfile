FROM --platform=$BUILDPLATFORM golang:alpine AS build-0

# Install git.
RUN apk update && \
    apk add --no-cache git=~2

# Set up working directory
WORKDIR /app

# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV GO111MODULE=on
ENV CGO_ENABLED=0

RUN go mod download -x

# Copy the rest of the project to ensure code changes doesnt trigger re-download of all deps
COPY . .

FROM --platform=$BUILDPLATFORM build-0 AS build-1

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
  go build -a -installsuffix cgo -o go-deploy_${TARGETOS}_${TARGETARCH} .


# Runner 
FROM alpine:3

# Set up the working directory
WORKDIR /go

ARG TARGETOS
ARG TARGETARCH

# Copy the binary from the builder stage
COPY --from=build-1 --chmod=777 /app/go-deploy_${TARGETOS}_${TARGETARCH} /usr/bin/go-deploy

# Copy the "index" folder
COPY --from=build-1 /app/index index

# Copy the "docs" folder
COPY --from=build-1 /app/docs docs

# Set environment variables and expose necessary port
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080

# Run the Go Gin binary
ENTRYPOINT ["/usr/bin/go-deploy"]

