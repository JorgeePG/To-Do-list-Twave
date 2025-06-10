run:
	go run ./cmd/todo

build:
	go build -o todo ./cmd/todo

test:
	go test ./...

docker:
	docker build -t todo-app .

sqlboiler:
	sqlboiler psql