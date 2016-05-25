default: build

build:
		go build -o bin/fusis

docker:
		go build -o bin/fusis && docker build -t fusis .

run:
	sudo bin/fusis balancer --single
