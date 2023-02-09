FROM golang:1.19 as builder

WORKDIR /go/src/github.com/rajatjindal/fermyon-cloud-preview
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go test ./... -cover
RUN CGO_ENABLED=0 GOOS=linux go build --ldflags "-s -w" -o fermyon-cloud-preview main.go

FROM alpine:3.17.1

WORKDIR /home/app

# Add non root user
RUN addgroup -S app && adduser app -S -G app
RUN chown app /home/app

USER app

COPY --from=builder /go/src/github.com/rajatjindal/fermyon-cloud-preview/fermyon-cloud-preview /usr/local/bin/

CMD ["fermyon-cloud-preview"]
