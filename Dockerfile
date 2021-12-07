FROM golang as build

WORKDIR /code

COPY . /code

RUN go build -o k8s-dns-controller *.go

FROM alpine:3.14

RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Seoul /etc/localtime && \
    echo "Asia/Seoul" > /etc/timezone

COPY --from=build /code/k8s-dns-controller /service/k8s-dns-controller
WORKDIR /service
CMD ["./k8s-dns-controller"]