package api

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/vince15dk/k8s-operator-ingress/pkg/data/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

const (
	secretName = "dnsplus-secret"
)

func generateHeader(h *http.Header) *http.Header {
	h.Set("Content-Type", "application/json")
	return h
}

func generateSecret(client kubernetes.Interface, namespace string) (*scheme.Secret, error) {
	s, err := client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	secret := scheme.Secret{
		AppKey:   string(s.Data["appKey"]),
		UserName: string(s.Data["userName"]),
	}
	return &secret, nil
}

func PostHandlerFunc(url string, body interface{}, header *http.Header) (*http.Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil{
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBytes))
	request.Header = *header
	client := http.Client{}
	return client.Do(request)
}
