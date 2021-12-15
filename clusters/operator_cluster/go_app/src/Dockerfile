FROM golang:1.17-alpine

RUN mkdir /app
RUN mkdir /app/static
RUN apk add --no-cache fortune

ADD . /app
WORKDIR /app

RUN go mod download
RUN go build -o main .

EXPOSE 8080
CMD ["/app/main"]
