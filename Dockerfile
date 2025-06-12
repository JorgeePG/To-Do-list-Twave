FROM golang:1.24.3

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN cd cmd/ && go build -o todo

EXPOSE 8080

CMD ["./todo"]
