FROM golang

WORKDIR /code

COPY . /code

RUN go build -o k8s-dns-controller *.go

ENTRYPOINT ["/code/k8s-dns-controller"]