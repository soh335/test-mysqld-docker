FROM golang:1.7-onbuild

RUN wget https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz -O /tmp/docker-latest.tgz \
        && tar -C /tmp -xvzf /tmp/docker-latest.tgz \
        && mv /tmp/docker/* /usr/local/bin/

RUN go get -t -v
