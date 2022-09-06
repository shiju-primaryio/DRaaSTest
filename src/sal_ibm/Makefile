#all: deps build
all: build

.PHONY: deps
deps:
	go get -u github.com/gin-gonic/gin
	go get github.com/IBM/ibm-cos-sdk-go/aws

.PHONY: build
#build: deps
build: 
	go build -o sal_ibm sal_rest_server.go ibm_api.go
clean:
	rm -rf sal_ibm  sal_logs
