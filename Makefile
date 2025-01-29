
all: web build
	@./bin/main

build:
	@go build -o ./bin/main cmd/api/*

web:
	@templ generate & tailwindcss -i ./public/input.css -o ./public/assets/css/styles.css

sshfile:
	@ssh -p 2222 localhost < ./banner.txt

scpfile: 
	@scp -P 2222 ./banner.txt localhost:

css:
	@tailwindcss -i ./public/assets/css/input.css -o ./public/assets/css/styles.css --watch

clean:
	@ssh-keygen -f "/home/ajpz/.ssh/known_hosts" -R "[localhost]:2222"
	
.PHONY: run build clean css
