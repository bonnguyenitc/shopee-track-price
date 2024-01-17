FROM golang:1.21.5

WORKDIR /usr/src/backend-app

RUN go install github.com/cosmtrek/air@latest

COPY . .
RUN go mod tidy