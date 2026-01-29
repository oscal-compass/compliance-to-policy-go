FROM alpine:3.23.3

RUN apk --no-cache add git yq

COPY install-kubectl.sh .

RUN ./install-kubectl.sh