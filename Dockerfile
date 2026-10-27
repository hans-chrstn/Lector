FROM node:24-alpine AS web
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS server
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /lector ./cmd/lector

FROM alpine:3.23
RUN apk add --no-cache ca-certificates && addgroup -g 10001 -S lector && adduser -u 10001 -S lector -G lector
WORKDIR /app
COPY --from=server /lector /app/lector
COPY --from=web /app/frontend/dist /app/public
RUN mkdir -p /app/data /app/plugins && chown -R lector:lector /app
USER lector
ENV LISTEN_ADDR=0.0.0.0:3000 DATA_DIR=/app/data PUBLIC_DIR=/app/public PLUGIN_DIR=/app/plugins
EXPOSE 3000
ENTRYPOINT ["/app/lector"]
