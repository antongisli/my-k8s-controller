# Use the official Go image with version 1.22.1 to create a build artifact.
FROM golang:1.22.1 as builder

# Copy local code to the container image.
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o my-k8s-controller .

# Use a Docker multi-stage build to create a lean production image.
FROM alpine:3
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/my-k8s-controller /my-k8s-controller

# Run the web service on container startup.
CMD ["/my-k8s-controller"]

