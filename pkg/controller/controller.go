package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/vince15dk/k8s-operator-ingress/pkg/api"
	ingressv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	netowrkingInformers "k8s.io/client-go/informers/extensions/v1beta1"
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
	secretName               = "dns-token"
	annotationConfigKey      = "nhn.cloud/dnsplus-config"
	annotationHost           = "nhn.cloud/dnsplus-hosts"
	annotationLoadBalancerIp = "nhn.cloud/dsnplus-loadbalancer-ip"
)

type Controller struct {
	client        kubernetes.Interface
	clusterSynced cache.InformerSynced
	lister        networkingLister.IngressLister
	wq            workqueue.RateLimitingInterface
	state         string
}

func NewController(client kubernetes.Interface, informer netowrkingInformers.IngressInformer) *Controller {
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
	for c.processNextItem() {}
}

func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		return false
	}
	defer c.wq.Done(item)
	defer c.wq.Forget(item)
	ctx := context.Background()

	if c.state == "update" {
		updateItem, ok := item.([2]interface{})
		if !ok {
			log.Printf("error %s", errors.New("item is not converted"))
			return false
		}

		oldObj := make([]interface{}, 0)
		newObj := make([]interface{}, 0)

		for _, r := range updateItem[0].(*ingressv1.Ingress).Spec.Rules {
			oldObj = append(oldObj, r.Host)
		}

		for _, r := range updateItem[1].(*ingressv1.Ingress).Spec.Rules {
			newObj = append(newObj, r.Host)
		}
		//
		//fmt.Println(oldObj)
		//fmt.Println(newObj)
		//fmt.Println("using golang-set library")
		//oldSet := set.NewSetFromSlice(oldObj)
		//newSet := set.NewSetFromSlice(newObj)
		//result1 := oldSet.Difference(newSet) // show deleted one
		//result2 := newSet.Difference(oldSet) // show added one
		//
		//fmt.Println(result1.ToSlice())
		//fmt.Println(result2.ToSlice())

	} else { // test
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
		// check if the object has been deleted from k8s cluster

		switch c.state {
		case "create":
			ingress, err := c.lister.Ingresses(ns).Get(name)
			if err != nil{
				log.Printf("error %s\n", err.Error())
				return false
			}
			b, _ := strconv.ParseBool(ingress.ObjectMeta.Annotations[annotationConfigKey])
			if b {
				annotationHosts := strings.Split(ingress.ObjectMeta.Annotations[annotationHost], ",")
				s := make([]string, 0)
				for _, ah := range annotationHosts {
					for _, h := range ingress.Spec.Rules {
						if ah == h.Host {
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
				go c.AddRecord(ns, name, ingress, annotationHosts, d)
			}

		case "delete":
			b, _ := strconv.ParseBool(item.(*ingressv1.Ingress).Annotations[annotationConfigKey])
			if b {
				aH := strings.Split(item.(*ingressv1.Ingress).Annotations[annotationHost], ",")
				ingressList, err := c.client.ExtensionsV1beta1().Ingresses("").List(ctx, metav1.ListOptions{})
				if err != nil {
					log.Printf("error %s\n", err.Error())
				}
				iL := make([]string, 0)
				for _, el := range ingressList.Items {
					ela := strings.Split(el.ObjectMeta.Annotations[annotationHost], ",")
					for _, elb := range ela {
						iL = append(iL, elb)
					}
				}
				dList := make([]string, 0)
			lo: for _, rh := range aH {
					for _, l := range iL {
						if rh == l {
							continue lo
						}
					}
					dList = append(dList, rh)
				}

				d := api.DnsHandler{
					Client:    c.client,
				}
				dnsList := d.ListDnsPlusZone(ns)
				fList := make([]string, 0)
				for n, id := range dnsList{
					for _, f := range dList{
						if strings.TrimSuffix(n, ".") == f{
							fList = append(fList, id)
						}
					}
				}

				if len(fList) > 0 {
					d.DeleteDnsPlusZone(ns, fList)
					log.Println("Deleting DnsPlusZone")
				}
			}
		}
	} //end
	return true
}

func (c *Controller) AddRecord(ns, name string, ingress *ingressv1.Ingress, aH []string, d api.DnsHandler) {
	lb, err := c.waitForIngressLB(ns, name)
	if err != nil {
		log.Printf("error %s, wating for ingress loadbalancer ip to be displayed", err.Error())
		//return false
	}
	ListRecords := make(map[int]string)
	for i, rH := range ingress.Spec.Rules {
		for _, h := range aH {
			if strings.Contains(rH.Host, h) {
				ListRecords[i] = fmt.Sprintf("%s.", rH.Host)
			}
		}
	}
	r := api.RecordSetHandler{
		Client:      c.client,
		ListRecords: ListRecords,
	}

	zoneList := d.ListDnsPlusZone(ns)
	r.CreateRecordSet(ns, lb, zoneList)
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
	//c.wq.Add(obj)
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
	c.wq.AddAfter(s, time.Second*2)
}
