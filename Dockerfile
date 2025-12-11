FROM golang:1.20-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o dsproxy ./cmd/dsproxy

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/dsproxy .
ENV DATABASE_URL=""
ENV REDIS_ADDR=""
EXPOSE 8080
CMD ["/app/dsproxy"]
