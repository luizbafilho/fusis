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
	sudo -E govendor test +local

PACKAGES = $(shell find ./ -type d -not -path '*/\.*' | grep -v /vendor)
ci:
	echo "mode: count" > coverage.txt
	$(foreach pkg,$(PACKAGES),\
		sudo -E go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage.txt;)
