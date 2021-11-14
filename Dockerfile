FROM golang:1.12rc1-alpine3.9 AS build

RUN apk --no-cache add ca-certificates
COPY . /go/src/github.com/miguelmota/go-rpc-provider-proxy
WORKDIR /go/src/github.com/miguelmota/go-rpc-provider-proxy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goproxy cmd/proxy/main.go

FROM scratch

WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /go/src/github.com/miguelmota/go-rpc-provider-proxy/goproxy .

EXPOSE 8000

CMD ["./goproxy"]
