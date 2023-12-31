FROM golang:latest

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o /go-forth

EXPOSE 8080

CMD [ "/go-forth" ]
