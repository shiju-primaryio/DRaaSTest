#all: deps build
all: build

.PHONY: deps
deps:
	go get -u github.com/gin-gonic/gin

.PHONY: build
#build: deps
build: 
	go build -o restord_rest_server restord_rest_server.go
clean:
	rm -rf restord_rest_server restord_logs
