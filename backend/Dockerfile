# ---- build stage ----
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# binary เล็ก, static link
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/api ./cmd/api

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Bangkok
COPY --from=builder /bin/api /bin/api
EXPOSE 8080
USER nobody
ENTRYPOINT ["/bin/api"]
