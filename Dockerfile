FROM debian:sid-slim

ENV KUBERNETES_VERSION=1.10.9

RUN apt-get update && \
        apt-get install -y gettext openssl curl tar ca-certificates git sudo && \
        curl -sSL https://get.docker.com/ | sh && \
        apt-get update && \
        apt-get upgrade -y && \
        apt-get install docker-ce && \
        apt-get clean && \
        rm -r /var/lib/apt/lists/* && \
        curl -L -o /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" && \
        chmod +x /usr/bin/kubectl && \
        kubectl version --client

WORKDIR /usr/local/bin

ADD ./* ./

ENV BUILD_TOOLS_PATH=/usr/local/bin

RUN source ./docker.sh
