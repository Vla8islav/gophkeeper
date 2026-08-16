FROM golang:1.26-bookworm AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG TARGETOS=linux
ARG TARGETARCH=amd64
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -o /build/gophkeeper-server ./cmd/gophkeeper-server

RUN mkdir -p /out/var/log/gophkeeper

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /build/gophkeeper-server /usr/local/bin/gophkeeper-server
COPY --from=builder /app/migrations /app/migrations
COPY --from=builder --chown=nonroot:nonroot /out/var/log/gophkeeper /var/log/gophkeeper

USER nonroot
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/gophkeeper-server"]
