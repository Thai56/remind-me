# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=sms
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build
build: 
		$(GOBUILD) -o$(BINARY_NAME) -v
test: 
		$(GOTEST) -v ./...
clean: 
		$(GOCLEAN)
		rm -f $(BINARY_NAME)
		rm -f $(BINARY_UNIX)
run:
		make clean
		$(GOBUILD) -v .
		./$(BINARY_NAME)
serve_static:		
		./stop_serving.sh
		docker build -f Dockerfile.web -t webserver-image:v1 .
		docker run -d -p 80:80 webserver-image:v1
sms_server:
	$(GOCMD) run sms_main.go
build_linux:
	env GOOS=linux GOARCH=arm go build -v github.com/user/sms
