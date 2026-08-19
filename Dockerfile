FROM oven/bun:1-alpine AS frontend
WORKDIR /app
ENV GIT_COMMIT=unknown
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile
COPY . .
RUN bun run build

FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/dist ./dist
RUN CGO_ENABLED=0 go build -o edna .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/edna .
EXPOSE 9325
ENTRYPOINT ["./edna"]
CMD ["-run-prod"]
