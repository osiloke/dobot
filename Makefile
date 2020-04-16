VERSION := 0.0.3
NAME := $(shell echo $${PWD\#\#*/})
TARGET := ./docker/$(NAME)
all: clean build image tag push
$(TARGET): 
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -ldflags="-X main.VERSION=$(VERSION) -X main.BUILD=$(shell git describe --always --long --dirty)" -o $(TARGET) github.com/osiloke/dobot
build: $(TARGET)
		@true
image:
	@docker build -t $(NAME):$(VERSION) ./docker
tag:
	@docker tag $(NAME):$(VERSION) rg.fr-par.scw.cloud/dostow/$(NAME):$(VERSION)
push:
	@docker push rg.fr-par.scw.cloud/dostow/$(NAME):$(VERSION)
clean:
	@rm -f $(TARGET)