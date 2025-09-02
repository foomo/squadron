FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN adduser -D -u 1001 -g 1001 squadron

COPY squadron /usr/bin/

USER squadron
WORKDIR /home/squadron

ENTRYPOINT ["squadron"]
