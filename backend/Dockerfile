FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/travel-map-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

USER app

COPY --from=builder /bin/travel-map-api /bin/travel-map-api

EXPOSE 8080

ENTRYPOINT ["/bin/travel-map-api"]
