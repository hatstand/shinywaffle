FROM golang:1.19-alpine AS builder

WORKDIR /go/src/app
COPY . .

RUN cd control/cmd && GOARCH=arm GOARM=6 go build -ldflags="-w -s"

FROM balenalib/raspberry-pi-alpine:3.12-run

COPY --from=builder /go/src/app/control/cmd/cmd /server
COPY control/cmd/config.textproto /
COPY control/cmd/*.html /

CMD /server -config /config.textproto -port 80 -logtostderr

EXPOSE 80 8082

HEALTHCHECK --interval=30s --timeout=30s --start-period=60s --retries=3 CMD curl http://localhost:80/healthz