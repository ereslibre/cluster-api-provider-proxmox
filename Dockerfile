FROM golang:1.12

WORKDIR /go/src/github.com/ereslibre/cluster-api-provider-proxmox
COPY . .

RUN make install
CMD ["manager"]
