FROM golang:1.21.5

WORKDIR /usr/src/backend-app

RUN apt update && apt -y upgrade

RUN go install github.com/cosmtrek/air@latest

COPY . .
RUN go mod tidy