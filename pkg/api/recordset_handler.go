package api

import (
	"fmt"
	"github.com/vince15dk/k8s-operator-ingress/pkg/data/scheme"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"strings"
)

type RecordSetHandler struct {
	Client      kubernetes.Interface
	ListRecords map[int]string
}

func (r *RecordSetHandler) GetIngressEndpoint() {

}

func (r *RecordSetHandler) CreateRecordSet(namespace, lb string, zoneList map[string]string) {
	// generate secret struct from k8s secret
	s, err := generateSecret(r.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return
	}

	// generate header for POST
	h := generateHeader(&http.Header{})
	for _, record := range r.ListRecords {
		for zoneName, zoneId := range zoneList {
			if strings.Contains(record, zoneName) {
				url := fmt.Sprintf("%s/%s/%s/%s/%s", baseUrl, s.AppKey, "zones", zoneId, "recordsets")
				rs := scheme.Records{
					RecordSet: scheme.Recordset{
						RecordsetName: record,
						RecordsetType: "A",
						RecordsetTTL:  86400,
						RecordList: scheme.RecordList{
							{RecordContent: lb,
								RecordDisabled: false},
						},
					},
				}
				_, err := PostHandlerFunc(url, rs, h)
				if err != nil {
					log.Printf("error %s\n", err.Error())
					return
				}
			}
		}
	}
}

func (r *RecordSetHandler) DeleteRecordSet() {

}

func (r *RecordSetHandler) UpdateRecordSet() {

}
