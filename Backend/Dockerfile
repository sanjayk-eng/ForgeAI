FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/agent ./cmd/agent

FROM alpine:3.22

RUN apk --no-cache add ca-certificates \
    && adduser -D -H -u 10001 appuser

COPY --from=build /out/agent /usr/local/bin/agent

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/agent"]