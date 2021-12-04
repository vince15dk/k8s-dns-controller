package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/kanisterio/kanister/pkg/poll"
	"github.com/vince15dk/k8s-operator-ingress/pkg/api"
	ingressv1 "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
	for c.processNextItem() {

	}
}

func (c *Controller) processNextItem() bool {
	item, shutDown := c.wq.Get()
	if shutDown {
		return false
	}
	defer c.wq.Forget(item)

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

		//fmt.Println(oldObj)
		//fmt.Println(newObj)
		//fmt.Println("using golang-set library")
		//oldSet := set.NewSetFromSlice(oldObj)
		//newSet := set.NewSetFromSlice(newObj)
		//result1 := oldSet.Difference(newSet) // show deleted one
		//result2 := newSet.Difference(oldSet) // show added one
		//
		//fmt.Println(result1)
		//fmt.Println(result2)


	}

	if c.state != "update" { // test
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
		ctx := context.Background()
		ingress, err := c.client.ExtensionsV1beta1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			if item.(*ingressv1.Ingress).Annotations[annotationConfigKey] == "true" {
				c.state = "delete"
			}
		}

		switch c.state {
		case "create":
			b, _ := strconv.ParseBool(ingress.ObjectMeta.Annotations[annotationConfigKey])
			if b {
				aH := strings.Split(ingress.ObjectMeta.Annotations[annotationHost], ",")
				m := make(map[int]string)
				for i, rH := range ingress.Spec.Rules {
					for _, h := range aH {
						if rH.Host == h {
							m[i] = fmt.Sprintf("%s.", rH.Host)
						}
					}
				}
				d := api.DnsHandler{
					Client:    c.client,
					ListHosts: m,
				}
				zoneList := d.ListDnsPlusZone(ns)

				d.CreateDnsPlusZone(ns, zoneList)

				// query dns plus to make sure ingress ip is provided
				lb, err := c.waitForIngressLB(ns, name)
				if err != nil {
					log.Printf("error %s, wating for ingress loadbalancer ip to be displayed", err.Error())
					return false
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

				zoneList = d.ListDnsPlusZone(ns)
				r.CreateRecordSet(ns, lb, zoneList)
			}

		case "delete":
			b, _ := strconv.ParseBool(item.(*ingressv1.Ingress).Annotations[annotationConfigKey])
			if b {
				aH := strings.Split(item.(*ingressv1.Ingress).Annotations[annotationHost], ",")
				m := make(map[int]string)
				for i, rH := range ingress.Spec.Rules {
					for _, h := range aH {
						if rH.Host == h {
							m[i] = fmt.Sprintf("%s.", rH.Host)
						}
					}
				}
				d := api.DnsHandler{
					Client:    c.client,
					ListHosts: m,
				}
				dnsList := d.ListDnsPlusZone(ns)
				d.DeleteDnsPlusZone(ns, dnsList)
				log.Println("Deleting DnsPlusZone")
			}
		case "update":
			fmt.Println("update called!")
		}
	} //end
	return true
}

func (c *Controller) waitForIngressLB(namespace, name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	lb := ""
	err := poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		ingress, err := c.client.ExtensionsV1beta1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			log.Printf("err err err %s\n", err.Error())
			return true, nil
		}
		for _, v := range ingress.Status.LoadBalancer.Ingress {
			if v.IP != "" {
				lb = v.IP
				return true, nil
			}
		}
		return false, nil
	})
	return lb, err
}

func checkIngressLister(c *Controller, namespace, name string) bool {
	ingress, err := c.lister.Ingresses(namespace).Get(name)
	if err != nil {
		log.Printf("error %s, Getting the ingress from lister", err.Error())
		return false
	}
	t := ingress.ObjectMeta.Annotations[annotationConfigKey]
	b, err := strconv.ParseBool(t)
	if err != nil {
		log.Printf("error %s, Failed to parse string to bool")
	}
	return b
}

func (c *Controller) handleIngressAdd(obj interface{}) {
	log.Println("Adding ingress handler is called")
	c.state = "create"
	c.wq.Add(obj)
}

func (c *Controller) handleIngressDelete(obj interface{}) {
	log.Println("Deleting ingress handler is called")
	c.state = "delete"
	c.wq.Add(obj)
}

func (c *Controller) handleIngressUpdate(old interface{}, new interface{}) {
	log.Println("Updating ingress handler is called")
	c.state = "update"
	s := [2]interface{}{
		old,
		new,
	}
	c.wq.Add(s)
}

