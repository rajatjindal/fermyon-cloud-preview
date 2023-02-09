FROM golang:1.19 as builder

WORKDIR /go/src/github.com/rajatjindal/fermyon-cloud-preview
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go test ./... -cover
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o fermyon-cloud-preview main.go

# spin
RUN wget https://github.com/fermyon/spin/releases/download/v0.8.0/spin-v0.8.0-linux-amd64.tar.gz && \
    tar -xvf spin-v0.8.0-linux-amd64.tar.gz && \
    ls -ltr && \
    mv spin /usr/local/bin/spin

FROM alpine:3.17.1

WORKDIR /home/app

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

USER app

COPY --from=builder /go/src/github.com/rajatjindal/fermyon-cloud-preview/fermyon-cloud-preview /usr/local/bin/
COPY --from=builder /usr/local/bin/spin /usr/local/bin/

CMD ["fermyon-cloud-preview"]
