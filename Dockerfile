# Step 1: Builder
FROM golang:1-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go test ./... && go build -ldflags="-s -w" -o /bin/app ./cmd/app

# Step 2: Final
FROM alpine:3
COPY --from=builder /app/config/config.yml /config/config.yml
COPY --from=builder /bin/app /go-service
CMD ["/go-service"]