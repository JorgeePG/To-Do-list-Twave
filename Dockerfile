FROM golang:1.24.3

WORKDIR /app/cmd

COPY . /app

RUN go mod download
RUN go build -o todo .

EXPOSE 8080

ENTRYPOINT ["./todo"]
CMD []
