FROM --platform=linux/amd64 golang:1.24.3-alpine AS builder
#ENV CGO_ENABLED=0
COPY . /flowerss
RUN apk add git make gcc libc-dev && \
    cd /flowerss && make build

# Image starts here
FROM --platform=linux/amd64 alpine
RUN mkdir -p /opt/flowerss/locales
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /flowerss/flowerss-bot /bin/
COPY --from=builder /flowerss/locales/ /opt/flowerss/locales/
RUN ls -alR /opt/flowerss/locales
VOLUME /root/.flowerss
WORKDIR /root/.flowerss
ENTRYPOINT ["/bin/flowerss-bot"]

