FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o image-converter .

FROM alpine:3.18
RUN apk add --no-cache imagemagick libgomp
COPY --from=builder /app/image-converter /usr/local/bin/
# Create a non-root user to run the application
RUN adduser -D appuser
USER appuser

ENTRYPOINT ["image-converter"]
