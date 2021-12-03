module github.com/vince15dk/k8s-operator-ingress

go 1.16

replace github.com/graymeta/stow => github.com/graymeta/stow v0.1.0

require (
	github.com/kanisterio/kanister v0.0.0-20211202074347-0b02f242b0e1
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/client-go v0.22.4
)
