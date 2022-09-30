FROM golang:1.19-alpine

RUN mkdir /app
RUN apk add --no-cache fortune

ADD . /app
WORKDIR /app

RUN go mod download
RUN go build -o main .

# TODO: copy from build env to scratch

EXPOSE 8080
CMD ["/app/main"]
