# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Copy and download dependencies first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fcitbot main.go

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /build/fcitbot .

# Copy any necessary files
COPY config.env /app/config.env
COPY cmds.txt /app/cmds.txt
COPY files/ /app/files/

# Create non-root user
RUN adduser -D appuser
USER appuser

# Command to run
ENTRYPOINT ["./fcitbot"]