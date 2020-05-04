#docker/dockerfile:latest
FROM golang:alpine3.11 as alpine

LABEL version="1.0.0"
LABEL maintainer="Sergey Sidorenko <carotage@mail.ru>"

RUN mkdir -p $GOPATH/src/incrementator
WORKDIR $GOPATH/src/incrementator
COPY . .
RUN go get -d -v
RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o $GOBIN/incrementator
FROM scratch
COPY --from=alpine $GOBIN/incrementator $GOBIN/incrementator
ENTRYPOINT ["$GOBIN/incrementator"]