.PHONY: run build lorem

run: ./cmd/main.go
	go run ./cmd/...

build: ./cmd/main.go
	go build -o gila ./cmd/...

lorem: build
	./gila testdata/lorem_ipsum.txt

short: build
	./gila testdata/short_lines.txt
