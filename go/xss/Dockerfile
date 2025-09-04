FROM golang:alpine as builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /cyberpoc

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o app main.go

FROM chromedp/headless-shell:latest

LABEL MAINTAINER="typesafe@qq.com"

WORKDIR /cyberpoc
COPY --from=builder /cyberpoc/app ./app
COPY --from=builder /cyberpoc/*.html ./

EXPOSE 80
ENTRYPOINT ./app