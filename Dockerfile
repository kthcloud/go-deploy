############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
RUN apk update && apk add --no-cache git=~2

# Set up working directory
WORKDIR $GOPATH/src/packages/goginapp/
COPY . .

# Fetch dependencies and build the binary
ENV GO111MODULE=on
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/main .

# Copy routers folder to /go/routers
RUN mkdir /go/routers
RUN cp -r /go/src/packages/goginapp/routers /go/routers

############################
# STEP 2 build a small image
############################
FROM alpine:3

# Set up the working directory
WORKDIR /go

# Copy the binary
COPY --from=builder /go/main /go/main

# Copy the routers folder
COPY --from=builder /go/routers /go/routers

# Set environment variables and expose necessary port
ENV PORT 8080
ENV GIN_MODE release
EXPOSE 8080

# Run the Go Gin binary
ENTRYPOINT ["/go/main"]