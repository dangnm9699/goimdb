FROM golang:latest

WORKDIR /crawl

COPY . .

RUN go build -mod=vendor main.go

CMD ["./main"]

