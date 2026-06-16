FROM alpine:3.24.1

RUN apk --no-cache add git yq

COPY install-kubectl.sh .

RUN ./install-kubectl.sh