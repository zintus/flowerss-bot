FROM --platform=linux/amd64 golang:1.18.6-alpine AS builder
#ENV CGO_ENABLED=0
COPY . /flowerss
RUN apk add git make gcc libc-dev && \
    cd /flowerss && make build

# Image starts here
FROM --platform=linux/amd64 alpine
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /flowerss/flowerss-bot /bin/
VOLUME /root/.flowerss
WORKDIR /root/.flowerss
ENTRYPOINT ["/bin/flowerss-bot"]

