# build stage
FROM golang:1.9.3-alpine3.7

ADD . /go/src/github.com/banzaicloud/preemption-exporter
WORKDIR /go/src/github.com/banzaicloud/preemption-exporter
RUN go build -o /bin/preemption-exporter .

FROM alpine:latest
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=0 /bin/preemption-exporter /bin
ENTRYPOINT ["/bin/preemption-exporter"]
