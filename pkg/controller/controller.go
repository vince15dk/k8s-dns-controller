package controller

import (
	"context"
	"fmt"
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
	secretName          = "dns-token"
	annotationConfigKey = "nhn.cloud/dnsplus-config"
	annotationHost      = "nhn.cloud/dnsplus-hosts"
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
			d.CreateDnsPlusZone(ns)
			log.Println("Creating DnsPlusZone")
		}

	case "delete":
		fmt.Println("inside delete")

	case "update":
		fmt.Println(ingress.ObjectMeta.Annotations[annotationConfigKey])
	}

	return true
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
	c.wq.Add(obj)
}

func (c *Controller) handleIngressUpdate(old interface{}, obj interface{}) {
	log.Println("Updating ingress handler is called")
	c.state = "update"
	c.wq.Add(obj)
}
