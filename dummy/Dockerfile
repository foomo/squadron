ARG IMAGE=golang:alpine
FROM $IMAGE

ARG IMAGE
COPY delve.sh delve.sh
RUN /bin/sh delve.sh $IMAGE

ENTRYPOINT ["/bin/sh", "-c", "while true; do echo dummy!; sleep 3600; done"]