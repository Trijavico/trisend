
run: build
	@./bin/main

build:
	@go build -o ./bin/main main.go

clean:
	@ssh-keygen -f "/home/ajpz/.ssh/known_hosts" -R "[localhost]:2222"
	

.PHONY: run build clean
