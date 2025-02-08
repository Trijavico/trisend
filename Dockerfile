FROM ubuntu:22.04
WORKDIR /app

RUN mkdir -p /app/templates
RUN mkdir -p /app/keys

COPY bin/main .

COPY templates/ /app/templates

RUN apt-get update && apt-get install -y ca-certificates openssh-client

RUN ssh-keygen -t rsa -f /app/keys/host

RUN chmod +x /app/main

EXPOSE 3000
EXPOSE 22

CMD ["/app/main"]
