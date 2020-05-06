#docker/dockerfile:latest
FROM golang:alpine3.11

LABEL version="1.0.0"
LABEL maintainer="Sergey Sidorenko <carotage@mail.ru>"
RUN mkdir -p /web/incrementator
WORKDIR /web/incrementator
COPY . .
RUN apk add git
RUN apk add --update gcc musl-dev
RUN go get -d -v
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o incrementator
CMD ["./incrementator"]
EXPOSE 8080
