# build stage
FROM golang:1.23 AS build-env
RUN mkdir -p /go/src/github.com/eumel8/prometheus-dashboard
WORKDIR /go/src/github.com/eumel8/prometheus-dashboard
COPY  . .
RUN useradd -u 10001 appuser
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o prometheus-dashboard

FROM alpine:latest
COPY --from=build-env /go/src/github.com/eumel8/navlinkswebhook/prometheus-dashboard .
COPY --from=build-env /etc/passwd /etc/passwd
USER appuser
ENTRYPOINT ["/prometheus-dashboard"]
