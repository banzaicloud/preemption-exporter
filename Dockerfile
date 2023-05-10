# build stage
FROM golang:1.20-alpine as builder
RUN apk add --no-cache ca-certificates

WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags '-extldflags "-static"' -o /bin/preemption-exporter .

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt \
     /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /bin/preemption-exporter /preemption-exporter

EXPOSE 9189
ENTRYPOINT ["/preemption-exporter"]
