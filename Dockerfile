FROM golang:1.12-alpine AS builder
RUN apk add git
ARG SRC_DIR=/src
ADD . $SRC_DIR
WORKDIR $SRC_DIR
RUN go build .

FROM alpine:latest
RUN apk add ca-certificates
COPY --from=builder /src/prom-config-controller /usr/bin/prom-config-controller
ENTRYPOINT ["/usr/bin/prom-config-controller"]
