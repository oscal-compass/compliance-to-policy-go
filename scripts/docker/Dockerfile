FROM alpine:3.23.4

RUN apk --no-cache add git yq

COPY install-kubectl.sh .

RUN ./install-kubectl.sh