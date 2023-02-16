.PHONY: run build

run: main.go
	go run main.go

build: main.go
	go build -o gila main.go
