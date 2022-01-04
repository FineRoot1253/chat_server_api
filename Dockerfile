FROM --platform=$BUILDPLATFORM golang:alpine as builder 
LABEL AUTHOR JunGeun Hong (gjhong1129@gmail.com)
RUN apk add ca-certificates && update-ca-certificates

WORKDIR /Users/reindeermacbook/img/chat_server
COPY . .


RUN go mod tidy \
    && go get -u -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$BUILDPLATFORM go build -a -ldflags '-s -w' -o main ./main.go

FROM --platform=$BUILDPLATFORM scratch


# RUN update-ca-certificates
COPY --from=builder /Users/reindeermacbook/img/chat_server /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 50000
CMD ["/main"]