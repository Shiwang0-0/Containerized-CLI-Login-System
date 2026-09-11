# Build stage
FROM golang:1.26.6-alpine AS builder

WORKDIR /app

# Cache dependency downloads separately
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 for a fully static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/cli-login ./cmd

# Runtime stage
FROM alpine:3.20

# tzdata: correct timezone handling for time.Now() / lockout timestamps
RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /app/bin/cli-login .
COPY migrations ./migrations

# Interactive CLI needs a real stdin/tty — set at `docker run`/compose level too
ENTRYPOINT ["./cli-login"]