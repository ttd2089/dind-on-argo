FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o app ./cmd/memhog

FROM gcr.io/distroless/static

COPY --from=builder /build/app /
USER 65532
ENTRYPOINT ["/app"]
