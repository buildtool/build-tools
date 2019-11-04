FROM debian:stretch-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    useradd -m -u 1001 vsts_VSTSContainer


COPY build push deploy kubecmd /usr/local/bin/
