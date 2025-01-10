# build stage
FROM golang:1.23 AS build-env
RUN mkdir -p /go/src/github.com/eumel8/prometheus-dashboard
WORKDIR /go/src/github.com/eumel8/prometheus-dashboard
COPY  . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -o prometheus-dashboard
# release stage
FROM alpine:latest
RUN adduser -u 10001 -h appuser -D appuser
WORKDIR /appuser
COPY --from=build-env /go/src/github.com/eumel8/prometheus-dashboard .
COPY --from=build-env /etc/passwd /etc/passwd
USER appuser
ENV PROMETHEUS_URL
ENV ALERTMANAGER_URL
ENTRYPOINT ["/appuser/prometheus-dashboard"]
