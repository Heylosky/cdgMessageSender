FROM golang:1.19 AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build
COPY . .
RUN go build -o messageSender .

FROM scratch
COPY --from=builder /build/messageSender /
COPY --from=builder /build/msgSender.log /
COPY --from=builder /build/MBconfig.json /
COPY --from=builder /build/server.crt /etc/ssl/certs/
ENTRYPOINT ["/messageSender"]