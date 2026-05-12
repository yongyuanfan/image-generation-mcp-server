FROM golang:1.25 AS builder

ENV GOPROXY https://goproxy.cn/

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/image-generation-mcp-server ./cmd/server

FROM alpine:3.22.4

WORKDIR /app

COPY --from=builder /out/image-generation-mcp-server /app/image-generation-mcp-server

EXPOSE 9101

ENTRYPOINT ["/app/image-generation-mcp-server"]
