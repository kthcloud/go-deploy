############################
# STEP 1 build executable binary
############################
FROM golang:1.20 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

############################
# STEP 2 build a small image
############################
FROM alpine:3

# Set the working directory in the container
WORKDIR /go

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the "index" folder
COPY --from=builder /app/index index

ENV PORT 8080
ENV GIN_MODE release

# Command to run the executable
ENTRYPOINT "./main"