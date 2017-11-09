//go:generate protoc --go_out=plugins=grpc:. -I$GOPATH/src -I$GOPATH/src/github.com/hatstand/shinywaffle/control -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis -I/usr/local/include --grpc-gateway_out=logtostderr=true:. control.proto service.proto
package control
