FROM golang:1.13 as go-build

RUN go get sigs.k8s.io/aws-iam-authenticator/cmd/aws-iam-authenticator

FROM debian:stretch-slim

RUN apt-get update && \
    apt-get install -y ca-certificates && \
    useradd -m -u 1001 vsts_VSTSContainer


COPY build push deploy kubecmd /usr/local/bin/
COPY --from=go-build /go/bin/aws-iam-authenticator /usr/local/bin/

