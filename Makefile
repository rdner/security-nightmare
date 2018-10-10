default: build

build:
	dep ensure -v && go build

run: build
	./security-nightmare
