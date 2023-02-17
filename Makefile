.PHONY: run build

run: main.go
	go run .

build: main.go
	go build -o gila .
