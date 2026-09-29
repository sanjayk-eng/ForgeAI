FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/agent ./cmd/agent

FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && adduser -D -H -u 10001 appuser

WORKDIR /app

COPY --from=build /out/agent ./agent
COPY migrations ./migrations

ENV HOST=0.0.0.0 \
    PORT=8080

EXPOSE 8080
USER appuser

ENTRYPOINT ["./agent"]
