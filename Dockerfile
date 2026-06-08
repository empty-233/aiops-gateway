FROM golang:1.25.8-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /aiops-gateway ./cmd/server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/configs /app/data \
    && chown -R nobody:nogroup /app/data

COPY --from=builder /aiops-gateway /app/aiops-gateway
COPY configs/config.example.yaml /app/configs/config.yaml

USER nobody

EXPOSE 8080

ENTRYPOINT ["/app/aiops-gateway"]
