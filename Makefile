.PHONY: run build lorem

run: main.go
	go run .

build: main.go
	go build -o gila .

lorem: build
	./gila testdata/lorem_ipsum.txt

short: build
	./gila testdata/short_lines.txt
