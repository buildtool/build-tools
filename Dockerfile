FROM frolvlad/alpine-glibc

ENV KUBERNETES_VERSION=1.10.9

RUN apk add -U docker findutils gettext openssl curl tar gzip bash ca-certificates git && \
        curl -L -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
        curl -L -O https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.28-r0/glibc-2.28-r0.apk && \
        apk add glibc-2.28-r0.apk && \
        rm glibc-2.28-r0.apk && \
        curl -L -o /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/v${KUBERNETES_VERSION}/bin/linux/amd64/kubectl" && \
        chmod +x /usr/bin/kubectl && \
        kubectl version --client

WORKDIR /usr/local/bin

ADD ./* ./

ENV BUILD_TOOLS_PATH=/usr/local/bin
