default: build

build:
	go build -race -o bin/fusis

run:
	sudo bin/fusis balancer --log-level=debug

docker:
	docker build -t fusis .

test:
	sudo -E go test $$(go list ./... | grep -v /vendor)

ci:
	./covertests.sh
