FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy go.mod and go.sum from backend folder
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./

RUN go build -o server .

FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/server .

ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
