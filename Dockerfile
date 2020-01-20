# Build the manager binary
FROM golang:1.13 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM ubuntu:bionic

RUN apt update && apt install -y curl git

RUN curl -Lo /usr/local/bin/tmplt https://github.com/mmlt/tool-tmplt/releases/download/v0.6.0/tmplt-v0.6.0-linux-amd64 \
 && chmod +x /usr/local/bin/tmplt \
 && curl -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl \
 && chmod +x /usr/local/bin/kubectl

WORKDIR /
COPY --from=builder /workspace/manager .

#TODO remove --groups sudo
RUN groupadd oprtr -g 1000 \
 && useradd --gid 1000 -u 1000 --groups sudo -s /bin/bash -m oprtr
USER 1000

ENTRYPOINT ["/manager"]
