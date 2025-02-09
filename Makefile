all: build
	@./bin/main

docker: build
	@docker build . -t trisend:latest

build: web
	@go build -o ./bin/main cmd/api/*

web:
	@templ generate & tailwindcss -i ./public/input.css -o ./public/assets/css/styles.css

css:
	@tailwindcss -i ./public/assets/css/input.css -o ./public/assets/css/styles.css --watch

sshkey:
	@if [ ! -d keys ]; then mkdir -p keys; fi
	@if [ -f keys/host ]; then echo "SSH key already exists in keys/ DIR"; \
		else ssh-keygen -t rsa -f keys/host; fi

sshfile:
	@ssh -p 2222 localhost banner.txt < ./internal/server/banner.txt

scpfile: 
	@scp -P 2222 ./internal/server/banner.txt localhost:

scpdir:
	@scp -P 2222 -r ./templates/ localhost:

clean:
	@ssh-keygen -f ~/.ssh/known_hosts -R "[localhost]:2222"
	
.PHONY: web docker build clean css
