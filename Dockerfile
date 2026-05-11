FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/image-generation-mcp-server ./cmd/server

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /out/image-generation-mcp-server /app/image-generation-mcp-server

EXPOSE 8080

ENTRYPOINT ["/app/image-generation-mcp-server"]
