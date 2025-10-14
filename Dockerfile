# MIT License
#
# Copyright (c) 2018 buildtool
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.

FROM amd64/golang:1.25.3@sha256:89abb76a5137ebbdccb6d993666b0561adc0718cccd827a95d6885c0c555e7dc as go-build

RUN go install sigs.k8s.io/aws-iam-authenticator/cmd/aws-iam-authenticator@v0.7.2

FROM debian:bookworm-20250520-slim@sha256:90522eeb7e5923ee2b871c639059537b30521272f10ca86fdbbbb2b75a8c40cd

RUN apt-get update && \
    apt-get install -y ca-certificates curl unzip && \
    useradd -m -u 1001 vsts_VSTSContainer

RUN cd /tmp &&  \
    curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip" && \
    unzip awscliv2.zip && \
    ./aws/install && \
    rm -rf aws && rm awscliv2.zip

COPY build push deploy kubecmd promote /usr/local/bin/
COPY --from=go-build /go/bin/aws-iam-authenticator /usr/local/bin/
