ARG BASE_IMAGE=nginx
ARG BASE_IMAGE_TAG=latest
FROM ${BASE_IMAGE}:${BASE_IMAGE_TAG}

COPY ./index.html /etc/nginx/templates/.
