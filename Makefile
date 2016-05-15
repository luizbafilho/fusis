default: build

build:
		go build -o bin/fusis && docker build -t fusis .
