run:
	go run ./cmd

build:
	go build -o todo ./cmd

test:
	go test ./...

docker:
	docker build -t todo-app .

sqlboiler:
	sqlite3 