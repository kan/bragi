FROM golang:1.22.4

WORKDIR /app
COPY . /app

RUN go build
VOLUME ["/go/pkg/mod"]

RUN go install github.com/air-verse/air@v1.49.0

CMD air