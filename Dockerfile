FROM debian:sid-slim

ENV KUBERNETES_VERSION=1.14.0

RUN groupadd -f -g 117 docker && \
        apt-get update && \
        apt-get install -y gettext openssl curl tar ca-certificates git sudo && \
        curl -sSL https://get.docker.com/ | sh && \
        apt-get update && \
        apt-get upgrade -y && \
        apt-get install docker-ce && \
        apt-get clean && \
        rm -r /var/lib/apt/lists/* && \
        curl -L -o /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" && \
        chmod +x /usr/bin/kubectl && \
        kubectl version --client && \
        useradd -m -u 1001 vsts_VSTSContainer && \
        usermod -a -G docker vsts_VSTSContainer

WORKDIR /usr/local/bin

COPY build deploy push service-setup ./

ENV BUILD_TOOLS_PATH=/usr/local/bin
