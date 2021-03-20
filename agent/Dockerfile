FROM golang:1.16 AS builder

WORKDIR /go/src/app
RUN wget https://github.com/grafana/agent/archive/refs/tags/v0.13.0.tar.gz
RUN tar -xvf v0.13.0.tar.gz
RUN cd agent-0.13.0 && GOARCH=arm GOARM=6 go build -ldflags="-w -s" ./cmd/agent

FROM balenalib/raspberry-pi-alpine:3.12-run

COPY --from=builder /go/src/app/agent-0.13.0/agent /agent
COPY agent-config.yml /agent-config.yml
CMD /agent -config.expand-env -config.file=/agent-config.yml