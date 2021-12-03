## generate code for clientset
execDir=~/go/src/k8s.io/code-generator
"${execDir}"/generate-groups.sh all github.com/vince15dk/k8s-operator-dnsplus/pkg/client github.com/vince15dk/k8s-operator-dnsplus/pkg/apis nhncloud.com:v1alpha1 --output-base "/Users/nhn/Desktop/Linux/Go/k8s-operator-dnsplus" --go-header-file "${execDir}"/hack/boilerplate.go.txt

"${execDir}"/generate-groups.sh all github.com/vince15dk/k8s-operator-dnsplus/pkg/client github.com/vince15dk/k8s-operator-dnsplus/pkg/apis nhncloud.com:v1alpha1 --go-header-file "${execDir}"/hack/boilerplate.go.txt

## generate manifests again after adding subresource `// +kubebuilder:subresource:status` to the resource
controller-gen paths=github.com/vince15dk/k8s-operator-dnsplus/pkg/apis/nhncloud.com/v1alpha1 crd:trivialVersions=true crd:crdVersions=v1 output:crd:artifacts:config=manifests

kl create secret generic dnsplus-secret --from-literal=appKey=KaRep2t4HVPw31TF --from-literal=userName=sukjoo.kim@nhn.com

# create dns zones
curl -X POST 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/zahH4VxBweLj4Jc8/zones' -H 'Content-Type: application/json' --data @data.json

# delete dns zones
curl -X DELETE 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/zahH4VxBweLj4Jc8/zones/async?zoneIdList=496fe8fb-4044-4641-b312-241de04fd108' -H 'Content-Type: application/json'

# add record set
curl -X POST 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/zahH4VxBweLj4Jc8/zones/496fe8fb-4044-4641-b312-241de04fd108/recordsets' -H 'Content-Type: application/json' --data @record.json

# delete record set
curl -X DELETE 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/zahH4VxBweLj4Jc8/zones/496fe8fb-4044-4641-b312-241de04fd108/recordsets?recordsetIdList=8cad1ac5-51f0-45c9-a4df-673f19ee5a1b' -H 'Content-Type: application/json'

# change record set
curl -X PUT 'https://api-dnsplus.cloud.toast.com/dnsplus/v1.0/appkeys/zahH4VxBweLj4Jc8/zones/496fe8fb-4044-4641-b312-241de04fd108/recordsets/26b1710d-8ffb-4a30-bee3-4002f06dcc3f' -H 'Content-Type: application/json' --data @record.json