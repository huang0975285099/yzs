FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server .

# Final image
FROM alpine:3.19
RUN apk add --no-cache tzdata ca-certificates
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/server .

EXPOSE 18881
CMD ["./server"]
