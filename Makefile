default: build

build:
	go build -o bin/fusis

run:
	sudo bin/fusis balancer --bootstrap

docker:
	docker build -t fusis .

deps:
	go get -u github.com/kardianos/govendor
	govendor add +external
	govendor sync

clean:
	sudo rm -rf /etc/fusis

test:
	sudo -E go test $$(go list ./... | grep -v /vendor)

ci:
	./covertests.sh
