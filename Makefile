.PHONY: build run clean tidy test docker sqlboiler

run:
	@echo "🔄 Ejecutando la aplicación..."
	cd cmd/ && go run . serve

list:
	@echo "🔄 Ejecutando la aplicación..."
	cd cmd/ && go run . list

build:
	@echo "🏗️ Compilando la aplicación..."
	go build -o todo ./cmd

test:
	@echo "🧪 Ejecutando pruebas..."
	go test ./...

docker:
	@echo "🐳 Construyendo imagen Docker..."
	docker build -t todo-app .

sqlboiler:
	@echo "⚙️ Ejecutando sqlboiler (actualmente placeholder)..."
	sqlite3

tidy:
	@echo "📦 Ordenando dependencias Go..."
	go mod tidy

clean:
	@echo "🧹 Limpiando archivos binarios..."
	del /Q todo-list-twave.exe 2>nul || true

full:
	@echo "📦 Ordenando dependencias..."
	go mod tidy
	@echo "🏗️ Compilando..."
	cd cmd/ && go build -o todo -buildvcs=false
	@echo "🧪 Ejecutando tests..."
	go test ./...
	@echo "🐳 Construyendo imagen Docker..."
	docker build -t todo-app .
	@echo "🚀 Ejecutando aplicación..."
	(cd cmd && go run . )  # Ejecutar en segundo plano
	@echo "✅ Proceso completo iniciado."
