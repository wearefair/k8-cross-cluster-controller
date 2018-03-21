FROM 889883130442.dkr.ecr.us-west-2.amazonaws.com/fair-images:golang-1.10.0-2 as builder

ARG GITHUB_OAUTH_TOKEN

WORKDIR /go/src/github.com/wearefair/k8-cross-cluster-controller
RUN mkdir /dist
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /dist/cross-cluster-controller-linux-x86

FROM scratch
COPY --from=builder dist/cross-cluster-controller-linux-x86 /bin/cross-cluster-controller
ADD https://curl.haxx.se/ca/cacert.pem /etc/ssl/ca-bundle.pem

CMD ["/bin/cross-cluster-controller"]
