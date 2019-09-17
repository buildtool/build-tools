FROM debian:stretch-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    useradd -m -u 1001 vsts_VSTSContainer

WORKDIR /usr/local/bin

COPY build deploy push ./

ENV BUILD_TOOLS_PATH=/usr/local/bin
