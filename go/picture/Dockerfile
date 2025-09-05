FROM golang:alpine as builder

ENV GO111MODULE=on

WORKDIR /cyberpoc

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o app main.go

FROM alpine:latest

LABEL MAINTAINER="typesafe@qq.com"

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add curl

WORKDIR /cyberpoc
COPY --from=builder /cyberpoc/app ./app

EXPOSE 80
ENTRYPOINT ./app