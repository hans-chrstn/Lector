FROM node:24-alpine AS frontend-builder

RUN npm install -g pnpm
WORKDIR /app

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY frontend/svelte.config.js frontend/tsconfig.json frontend/vite.config.ts ./
COPY frontend/src ./src
COPY frontend/static ./static

RUN pnpm build

FROM golang:1.26-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY cmd/ ./cmd/

RUN CGO_ENABLED=1 GOOS=linux go build -v -o lector cmd/lector/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates musl su-exec

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

RUN mkdir -p /app/data /app/uploads /app/exports /app/plugins && \
    chown -R appuser:appgroup /app/data /app/uploads /app/exports /app/plugins

COPY --from=backend-builder /app/lector .

COPY --from=frontend-builder /app/build ./public

COPY plugins/ ./plugins/

RUN chown -R appuser:appgroup /app/plugins

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

EXPOSE 3000

ENV NODE_ENV=production
ENV DATABASE_PATH=/app/data/lector.db

ENTRYPOINT ["/app/entrypoint.sh"]

CMD ["./lector"]
