FROM golang:1.23-bookworm AS builder

WORKDIR /src

# Download GO modules
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY *.go .
COPY pkg/ ./pkg/
COPY internal/ ./internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app

# Run the app on a basic scratch image
FROM gcr.io/distroless/static-debian12 AS release

WORKDIR /

# Copy the binary
ENV GIN_MODE=release
COPY --from=builder /app /app

EXPOSE 3000

# LLEEEEEXXXXXXXXXX
ENTRYPOINT ["/app"]