package api

import (
	"encoding/json"
	"fmt"
	"github.com/vince15dk/k8s-operator-ingress/pkg/data/scheme"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"strings"
)

const (
	baseUrl = "https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys"
)

type DnsHandler struct {
	Client    kubernetes.Interface
	ListHosts []string
	CachedHost map[string]string
}

func (d *DnsHandler) CreateDnsPlusZone(namespace string, zoneList map[string]string) {
	// generate secret struct from k8s secret
	s, err := generateSecret(d.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return
	}
	// generate header for POST
	h := generateHeader(&http.Header{})
	// Create DnsPlus
	url := fmt.Sprintf("%s/%s/%s", baseUrl, s.AppKey, "zones")
	addHosts := make([]string, 0)
	lo: for _, record := range d.ListHosts {
		for zoneName, _ := range zoneList {
			if strings.Contains(record, zoneName) {
				continue lo
			}
		}
		addHosts = append(addHosts, record)
	}
	fmt.Println(addHosts)

	for _, host := range addHosts {
		zone := scheme.DnsZone{
			Zone: scheme.Zone{
				ZoneName:    host,
				Description: fmt.Sprintf("Generated by k8s api"),
			},
		}
		_, err := PostHandlerFunc(url, zone, h)
		if err != nil {
			log.Printf("error %s", err.Error())
			return
		}
	}
}

func (d *DnsHandler) DeleteDnsPlusZone(namespace string) {
	// generate secret struct from k8s secret
	s, err := generateSecret(d.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return
	}
	// generate header for POST
	h := generateHeader(&http.Header{})

	// Delete Zones
	for _, zoneId := range d.ListHosts{
		url := fmt.Sprintf("%s/%s/%s=%s", baseUrl, s.AppKey, "zones/async?zoneIdList", zoneId)
		_, err := DeleteHandleFunc(url, h)
		if err != nil {
			log.Printf("error %s\n", err.Error())
			return
		}
	}
}

func (d *DnsHandler) ListDnsPlusZone(namespace string) map[string]string {
	// generate secret struct from k8s secret
	s, err := generateSecret(d.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return nil
	}
	// generate header for POST
	h := generateHeader(&http.Header{})
	// Create DnsPlus
	url := fmt.Sprintf("%s/%s/%s", baseUrl, s.AppKey, "zones")

	resp, err := ListHandlerFunc(url, h)
	if err != nil {
		log.Printf("error %s\n", err.Error())
		return nil
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error %s\n", err.Error())
		return nil
	}
	defer resp.Body.Close()
	dnsList := &scheme.DnsZoneList{
		TotalCount: 0,
		ZoneList:   scheme.ZoneList{},
	}
	err = json.Unmarshal(bytes, dnsList)
	m := make(map[string]string)
	for _, dns := range dnsList.ZoneList {
		m[dns.ZoneName] = dns.ZoneID
	}
	return m
}

