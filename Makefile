
run: build
	@./bin/main

build:
	@go build -o ./bin/main main.go

sshfile:
	@ssh -p 2222 localhost < ./banner.txt

scpfile: 
	@scp -P 2222 ./banner.txt localhost:

css:
	@tailwindcss -i ./assets/css/input.css -o ./assets/css/styles.css --watch

clean:
	@ssh-keygen -f "/home/ajpz/.ssh/known_hosts" -R "[localhost]:2222"
	
.PHONY: run build clean css
