## Ingress DNS Operator

## Test API
### list dns zones
```
curl -X GET 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}}/zones'
```

### create dns zones
```
curl -X POST 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones' -H 'Content-Type: application/json' --data @data.json
```

### delete dns zones
```
curl -X DELETE 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones/async?zoneIdList={zoneId}}' -H 'Content-Type: application/json'
```

### list record set
```
curl -X GET 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones/{zoneId}/recordsets'
```

### add record set
```
curl -X POST 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones/{zoneId}/recordsets' -H 'Content-Type: application/json' --data @record.json
```

### delete record set
```
curl -X DELETE 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones/{zoneId}/recordsets?recordsetIdList={recoredsetId}' -H 'Content-Type: application/json'
```

### change record set
```
curl -X PUT 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/{appkey}/zones/{zoneId}/recordsets/{recoredsetId}' -H 'Content-Type: application/json' --data @record.json
```