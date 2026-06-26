# ---- build stage ----
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/migrate /app/migrate
COPY migrations /app/migrations
# O runner de migrations procura ./migrations a partir do WORKDIR (/app).
EXPOSE 8080
ENTRYPOINT ["/app/api"]
