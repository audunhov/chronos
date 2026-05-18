# Backend build stage
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build the actual API entrypoint
RUN go build -o main ./cmd/api/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=backend-builder /app/main .

EXPOSE 8080
CMD ["./main"]
