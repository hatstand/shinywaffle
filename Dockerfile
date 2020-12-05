FROM golang:1.15-alpine

WORKDIR /go/src/app
COPY . .

RUN cd control/cmd && go build

CMD control/cmd/cmd -config control/cmd/config.textproto

EXPOSE 8081 8082