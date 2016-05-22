default: build

build:
		go build -o bin/fusis

restore:
	go get -u github.com/kardianos/govendor
	govendor add +external
	govendor sync
