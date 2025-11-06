# stage 1
FROM golang:1.14-alpine as build
ARG TZ='Asia/Ho_Chi_Minh'
RUN apk --update add make git

WORKDIR /app

ENV GO111MODULE=on
ENV GOPRIVATE=gitlab.com
ENV GOPROXY=direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . .

RUN make engine


# stage 2
FROM alpine:3.12
RUN apk --update add ca-certificates tzdata

ARG TZ='Asia/Ho_Chi_Minh'

RUN echo $TZ >  /etc/timezone

WORKDIR /app

COPY --from=build /app/engine .
COPY --from=build /app/config.json .
