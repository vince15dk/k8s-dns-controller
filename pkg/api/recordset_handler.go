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

type RecordSetHandler struct {
	Client      kubernetes.Interface
	ListRecords []string
	ListZones   []string
}

func (r *RecordSetHandler) ListRecordSet(namespace, zoneId string) map[string]string {
	// generate secret struct from k8s secret
	s, err := generateSecret(r.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return nil
	}
	h := generateHeader(&http.Header{})
	url := fmt.Sprintf("%s/%s/%s/%s/%s", baseUrl, s.AppKey, "zones", zoneId, "recordsets")
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
	recordList := &scheme.RecordSetLists{
		TotalCount:    0,
		RecordsetList: []scheme.RecordSetList{},
	}
	err = json.Unmarshal(bytes, recordList)
	m := make(map[string]string)
	for _, record := range recordList.RecordsetList {
		if record.RecordsetType == "A"{
			m[record.RecordsetName] = record.RecordsetID
		}
	}
	return m
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
		record = fmt.Sprintf("%s.", record)
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

func (r *RecordSetHandler) DeleteRecordSet(namespace, zoneId string) {
	// generate secret struct from k8s secret
	s, err := generateSecret(r.Client, namespace)
	if err != nil {
		log.Printf("error %s", err.Error())
		return
	}
	// generate header for POST
	h := generateHeader(&http.Header{})

	// Delete Records
	for _, recordSetId := range r.ListRecords {
		url := fmt.Sprintf("%s/%s/%s/%s/%s=%s", baseUrl, s.AppKey, "zones", zoneId, "recordsets?recordsetIdList", recordSetId)
		_, err := DeleteHandleFunc(url, h)
		if err != nil {
			log.Printf("error %s\n", err.Error())
			return
		}
	}
}