.PHONY: run build lorem short war

run: ./cmd/main.go
	go run ./cmd/...

build: ./cmd/main.go
	go build -o gila ./cmd/...

test:
	go test -race ./...

lorem: build
	./gila testdata/lorem_ipsum.txt

short: build
	./gila testdata/short_lines.txt

war: build
	./gila testdata/war_and_peace.txt
