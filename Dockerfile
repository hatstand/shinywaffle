FROM golang:1.15

WORKDIR /go/src/app
COPY . .

RUN cd control/cmd && go build

CMD control/cmd/cmd -config control/cmd/config.textproto
