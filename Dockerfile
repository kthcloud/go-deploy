############################
# STEP 1 build executable binary
############################
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
# Install git.
RUN apk update && apk add --no-cache git=~2

# Set up working directory
WORKDIR /app
# Copy go.mod and go.sum separately so we only invalidate the downloading layers if we need to
COPY go.mod go.sum ./

# Fetch dependencies and build the binary
ENV GO111MODULE=on
RUN go mod download

# Copy the rest of the project to ensure code changes doesnt trigger re-download of all deps
COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$GOARCH go build -a -installsuffix cgo -o main .


############################
# STEP 2 build a small image
############################
FROM alpine:3

# Set up the working directory
WORKDIR /go

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the "index" folder
COPY --from=builder /app/index index

# Copy the "docs" folder
COPY --from=builder /app/docs docs

# Set environment variables and expose necessary port
ENV PORT=8080
ENV GIN_MODE=release
EXPOSE 8080

# Run the Go Gin binary
ENTRYPOINT ["./main"]

