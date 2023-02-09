FROM golang:1.19 as builder

# spin
RUN wget https://github.com/fermyon/spin/releases/download/v0.8.0/spin-v0.8.0-linux-amd64.tar.gz && \
    tar -xvf spin-v0.8.0-linux-amd64.tar.gz && \
    ls -ltr && \
    mv spin /usr/local/bin/spin

WORKDIR /go/src/github.com/rajatjindal/fermyon-cloud-preview
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go test ./... -cover
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o fermyon-cloud-preview main.go



FROM ubuntu:22.04

# Add non root user
RUN adduser app --disabled-password --gecos ""
USER app
WORKDIR /home/app

COPY --from=builder /go/src/github.com/rajatjindal/fermyon-cloud-preview/fermyon-cloud-preview /usr/local/bin/
COPY --from=builder /usr/local/bin/spin /usr/local/bin/

CMD ["fermyon-cloud-preview"]
