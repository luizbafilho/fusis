default: build

build:
	GOOS=linux go build -o bin/fusis

run:
	GOOS=linux sudo bin/fusis balancer --log-level=debug

docker:
	docker build -t fusis .

test:
	GOOS=linux sudo -E go test -race $$(go list ./... | grep -v /vendor)

ci:
	./covertests.sh
