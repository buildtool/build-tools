FROM golang:1.19.0 as go-build

RUN go install sigs.k8s.io/aws-iam-authenticator/cmd/aws-iam-authenticator@latest

FROM debian:bullseye-20220328-slim

RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    useradd -m -u 1001 vsts_VSTSContainer


COPY build push deploy kubecmd /usr/local/bin/
COPY --from=go-build /go/bin/aws-iam-authenticator /usr/local/bin/

