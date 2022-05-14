.PHONY: all update-modules

all:
	env GO111MODULE=on go build -v . ./_example/simple/...

update-modules:
	env GO111MODULE=on go get -u
	env GO111MODULE=on go mod tidy -v
