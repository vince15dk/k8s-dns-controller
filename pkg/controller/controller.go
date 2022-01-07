package controller

import (
	"context"
	"errors"
	"fmt"
	set "github.com/deckarep/golang-set"
	"github.com/vince15dk/k8s-operator-ingress/pkg/api"
	ingressv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	networkingInformers "k8s.io/client-go/informers/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	networkingLister "k8s.io/client-go/listers/extensions/v1beta1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"strconv"
	"strings"
	"time"
)

const (
	secretName          = "dns-token"
	annotationConfigKey = "nhn.cloud/dnsplus-config"
	annotationHost      = "nhn.cloud/dnsplus-hosts"
	annotationStaticIp  = "nhn.cloud/dsnplus-static-ip"
)

type Controller struct {
	client        kubernetes.Interface
	clusterSynced cache.InformerSynced
	lister        networkingLister.IngressLister
	wq            workqueue.RateLimitingInterface
	state         string
}

func NewController(client kubernetes.Interface, informer networkingInformers.IngressInformer) *Controller {
	c := &Controller{
		client:        client,
		clusterSynced: informer.Informer().HasSynced,
		lister:        informer.Lister(),
		wq:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress"),
	}

	informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.handleIngressAdd,
			DeleteFunc: c.handleIngressDelete,
			UpdateFunc: c.handleIngressUpdate,
		})
	return c
}

func (c *Controller) Run(ch chan struct{}) error {
	if ok := cache.WaitForCacheSync(ch, c.clusterSynced); !ok {
		log.Println("cache was not synced")
	}

	go wait.Until(c.worker, time.Second, ch)
	<-ch
	return nil
}

func (c *Controller) worker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		return false
	}
	defer c.wq.Done(item)
	defer c.wq.Forget(item)

	if c.state == "update" {
		updateItem, ok := item.([2]interface{})
		if !ok {
			log.Printf("error %s", errors.New("item can not be converted"))
			return false
		}

		fmt.Println("this is called for update")
		fmt.Println("old")
		fmt.Println(updateItem[0].(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost])
		fmt.Println("new")
		fmt.Println(updateItem[1].(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost])

		ns := updateItem[1].(*ingressv1.Ingress).Namespace
		name := updateItem[1].(*ingressv1.Ingress).Name

		// rules
		oldObjRules := make([]interface{}, 0)
		newObjRules := make([]interface{}, 0)

		for _, r := range updateItem[0].(*ingressv1.Ingress).Spec.Rules {
			oldObjRules = append(oldObjRules, r.Host)
		}

		for _, r := range updateItem[1].(*ingressv1.Ingress).Spec.Rules {
			newObjRules = append(newObjRules, r.Host)
		}

		oldSetRules := set.NewSetFromSlice(oldObjRules)
		newSetRules := set.NewSetFromSlice(newObjRules)
		resultRules1 := oldSetRules.Difference(newSetRules) // show deleted one
		resultRules2 := newSetRules.Difference(oldSetRules) // show added one

		// hosts
		oldObjHosts := make([]interface{}, 0)
		newObjHosts := make([]interface{}, 0)

		for _, r := range strings.Split(updateItem[0].(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost], ",") {
			oldObjHosts = append(oldObjHosts, r)
		}

		for _, r := range strings.Split(updateItem[1].(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost], ",") {
			newObjHosts = append(newObjHosts, r)
		}

		oldSetHosts := set.NewSetFromSlice(oldObjHosts)
		newSetHosts := set.NewSetFromSlice(newObjHosts)
		resultHosts1 := oldSetHosts.Difference(newSetHosts) // show deleted one
		resultHosts2 := newSetHosts.Difference(oldSetHosts) // show added one

		if len(resultHosts1.ToSlice()) > 0 {
			s := make([]string, 0)
			for _, v := range resultHosts1.ToSlice() {
				s = append(s, v.(string))
			}
			deleteHosts := strings.Join(s, ",")
			updateItem[0].(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost] = deleteHosts
			c.handleIngressDelete(updateItem[0])
		}
		if len(resultHosts2.ToSlice()) > 0 {
			c.handleIngressAdd(updateItem[1])
		}
		if len(resultRules2.ToSlice()) > 0 {

			ingress, err := c.lister.Ingresses(ns).Get(name)
			if err != nil {
				log.Printf("error %s\n", err.Error())
				return false
			}
			// s is dns zone lists to be created
			d := api.DnsHandler{
				Client: c.client,
			}

			// query dns plus to make sure ingress ip is provided
			go c.AddRecord(ns, name, ingress, d, resultRules2.ToSlice())
		}
		// delete record
		if len(resultRules1.ToSlice()) > 0 {
			r := api.RecordSetHandler{
				Client: c.client,
			}
			d := api.DnsHandler{
				Client: c.client,
			}

			recordList := make([]string, 0)
			for _, v := range resultRules1.ToSlice() {
				recordList = append(recordList, v.(string))
			}

			dnsList := d.ListDnsPlusZone(ns)
			for _, v := range dnsList {
				r.ListZones = append(r.ListZones, v)
			}
			r.ListRecords = []string{}
			for _, zid := range r.ListZones {
				for rn, rd := range r.ListRecordSet(ns, zid) {
					for _, v := range recordList {
						if strings.TrimSuffix(rn, ".") == v {
							r.ListRecords = append(r.ListRecords, rd)
						}
					}
				}
				r.DeleteRecordSet(ns, zid)
			}
		}

	} else {
		key, err := cache.MetaNamespaceKeyFunc(item)
		if err != nil {
			log.Printf("error %s called Namespace key func on cache for item", err.Error())
			return false
		}
		ns, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			log.Printf("error %s, Getting the namespace from lister", err.Error())
			return false
		}

		switch c.state {
		case "create":
			ingress, err := c.lister.Ingresses(ns).Get(name)
			if err != nil {
				log.Printf("error %s\n", err.Error())
				return false
			}
			b, _ := strconv.ParseBool(ingress.ObjectMeta.Annotations[annotationConfigKey])
			if b {
				annotationHosts := strings.Split(ingress.ObjectMeta.Annotations[annotationHost], ",")
				s := make([]string, 0)
				for _, ah := range annotationHosts {
					for _, h := range ingress.Spec.Rules {
						if strings.Contains(h.Host, ah) {
							s = append(s, fmt.Sprintf("%s.", ah))
						}
					}
				}

				// s is dns zone lists to be created
				d := api.DnsHandler{
					Client:    c.client,
					ListHosts: s,
				}

				d.CreateDnsPlusZone(ns, d.ListDnsPlusZone(ns))

				// query dns plus to make sure ingress ip is provided
				go c.AddRecord(ns, name, ingress, d, nil)
			}

		case "delete":
			b, _ := strconv.ParseBool(item.(*ingressv1.Ingress).ObjectMeta.Annotations[annotationConfigKey])
			if b {
				annotationHosts := strings.Split(item.(*ingressv1.Ingress).ObjectMeta.Annotations[annotationHost], ",")
				ingressList, err := c.lister.List(labels.Everything())
				if err != nil {
					log.Printf("error %s\n", err.Error())
				}
				iL := make([]string, 0)
				for _, el := range ingressList {
					iL = append(iL, strings.Split(el.ObjectMeta.Annotations[annotationHost], ",")...)
				}

				zoneList := make([]string, 0)
				zoneNotDeletedList := make([]string, 0)
			lo:
				for _, rh := range annotationHosts {
					for _, l := range iL {
						if rh == l {
							zoneNotDeletedList = append(zoneNotDeletedList, l)
							continue lo
						}
					}
					zoneList = append(zoneList, rh)
				}
				d := api.DnsHandler{
					Client: c.client,
				}
				r := api.RecordSetHandler{
					Client: c.client,
				}

				dnsList := d.ListDnsPlusZone(ns)

				// dnsList zoneName: zoneId
				// to get id from dnsList
				for n, id := range dnsList {
					for _, f := range zoneList {
						if strings.TrimSuffix(n, ".") == f {
							d.ListHosts = append(d.ListHosts, id)
						}
					}
				}

				if len(d.ListHosts) > 0 {
					d.DeleteDnsPlusZone(ns)
					log.Println("Deleting DnsPlusZone")
				}

				for n, id := range dnsList {
					for _, f := range zoneNotDeletedList {
						if strings.TrimSuffix(n, ".") == f {
							r.ListZones = append(r.ListZones, id)
						}
					}
				}

				if len(zoneNotDeletedList) > 0 {
					r.ListRecords = []string{}
					for _, zid := range r.ListZones {
						for rn, rd := range r.ListRecordSet(ns, zid) {
							for _, v := range item.(*ingressv1.Ingress).Spec.Rules {
								if strings.TrimSuffix(rn, ".") == v.Host {
									r.ListRecords = append(r.ListRecords, rd)
								}
							}
						}
						r.DeleteRecordSet(ns, zid)
					}
				}
			}
		}
	} //end
	return true
}

func (c *Controller) AddRecord(ns, name string, ingress *ingressv1.Ingress, d api.DnsHandler, addedRecord []interface{}) {
	lb := ""
	if ingress.ObjectMeta.Annotations[annotationStaticIp] == "" {
		ingressLB, err := c.waitForIngressLB(ns, name)
		if err != nil {
			log.Printf("error %s, wating for ingress loadbalancer ip to be displayed", err.Error())
			return
		}
		lb = ingressLB
	} else {
		staticLB := ingress.ObjectMeta.Annotations[annotationStaticIp]
		lb = staticLB
	}
	recordList := make([]string, 0)
	if len(addedRecord) > 0 {
		for _, v := range addedRecord {
			recordList = append(recordList, v.(string))
		}
	} else {
		for _, v := range ingress.Spec.Rules {
			recordList = append(recordList, v.Host)
		}
	}
	r := api.RecordSetHandler{
		Client:      c.client,
		ListRecords: recordList,
	}
	r.CreateRecordSet(ns, lb, d.ListDnsPlusZone(ns))
}

func (c *Controller) waitForIngressLB(namespace, name string) (string, error) {
	count := 0
	for {
		ingress, err := c.client.ExtensionsV1beta1().Ingresses(namespace).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			log.Printf("err %s\n", err.Error())
			break
		}
		if count == 36 {
			break
		}
		for _, v := range ingress.Status.LoadBalancer.Ingress {
			if v.IP != "" {
				lb := v.IP
				return lb, nil
			}
		}
		count++
		time.Sleep(time.Second * 5)
	}
	return "", errors.New("unable to fetch ingress lb ip")
}

func (c *Controller) handleIngressAdd(obj interface{}) {
	log.Println("Adding ingress handler is called")
	c.state = "create"
	c.wq.AddAfter(obj, time.Second*2)
}

func (c *Controller) handleIngressDelete(obj interface{}) {
	log.Println("Deleting ingress handler is called")
	c.state = "delete"
	c.wq.AddAfter(obj, time.Second*2)
}

func (c *Controller) handleIngressUpdate(old interface{}, new interface{}) {
	log.Println("Updating ingress handler is called")
	c.state = "update"
	s := [2]interface{}{
		old,
		new,
	}
	c.wq.AddRateLimited(s)
}
