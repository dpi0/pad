FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN adduser -D -g '' -u 10001 paduser \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o pad main.go \
    && touch pad.txt \
    && chown 10001:10001 pad.txt

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder --chown=10001:10001 /app/static /static
COPY --from=builder --chown=10001:10001 /app/pad /pad
COPY --from=builder --chown=10001:10001 /app/pad.txt /pad.txt

ENV PORT=8080 \
    STATIC_DIR=/static \
    DATA_FILE=/data/pad.txt

USER 10001

EXPOSE 8080

ENTRYPOINT ["/pad"]
