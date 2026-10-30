# Stage 1: Build tracker and web studio
FROM node:22-alpine AS web-builder
WORKDIR /app

# Build tracker
COPY tracker/package*.json ./tracker/
RUN cd tracker && npm install
COPY tracker/ ./tracker/
RUN cd tracker && npm run build

# Build web dashboard
COPY web/package*.json ./web/
RUN cd web && npm install
COPY web/ ./web/
RUN cd web && npm run build

# Stage 2: Build single Go binary
FROM golang:1.23-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Copy built static assets from stage 1 into Go embed directory
COPY --from=web-builder /app/internal/server/dist ./internal/server/dist

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o /app/microhog cmd/microhog/main.go

# Stage 3: Minimal runtime image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=go-builder /app/microhog /usr/local/bin/microhog

EXPOSE 8080
VOLUME ["/data"]

ENV PORT=8080
ENV MICROHOG_DB=/data/microhog.db

ENTRYPOINT ["microhog"]
CMD ["serve", "--port", "8080", "--db", "/data/microhog.db"]
