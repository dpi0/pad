FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN adduser -D -g '' -u 1000 paduser \
    && mkdir -p /data \
    && chown -R 1000:1000 /app /data \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pad main.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder --chown=1000:1000 /app/static /static
COPY --from=builder --chown=1000:1000 /app/pad /pad
COPY --from=builder --chown=1000:1000 /data /data

ENV PORT=8080 \
    STATIC_DIR=/static \
    DATA_FILE=/data/pad.txt

USER 1000

EXPOSE 8080

ENTRYPOINT ["/pad"]
