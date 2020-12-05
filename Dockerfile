FROM golang:1.15-alpine AS builder

WORKDIR /go/src/app
COPY . .

RUN cd control/cmd && GOARCH=arm GOARM=6 go build -ldflags="-w -s"

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/app/control/cmd/cmd /server
COPY control/cmd/config.textproto /

CMD /server -config /config.textproto -port 80

EXPOSE 80 8082