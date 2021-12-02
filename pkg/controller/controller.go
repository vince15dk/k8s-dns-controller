package controller

import (

	"k8s.io/apimachinery/pkg/util/wait"
	netowrkingInformers "k8s.io/client-go/informers/networking/v1"
	"k8s.io/client-go/kubernetes"
	networkingLister "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

var (
	secretName = "dns-token"
)

type Controller struct {
	client        kubernetes.Interface
	clusterSynced cache.InformerSynced
	lister        networkingLister.IngressLister
	wq     workqueue.RateLimitingInterface
	state         string
}

func NewController(client kubernetes.Interface, informer netowrkingInformers.IngressInformer) *Controller {
	c := &Controller{
		client:        client,
		clusterSynced: informer.Informer().HasSynced,
		lister:        informer.Lister(),
		wq:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ingress"),
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
	log.Println("creating ingress")
	log.Println(ns, name)
	switch c.state{
	case "create":
		ingress, err := c.lister.Ingresses(ns).Get(name)
		if err != nil{
			log.Printf("error %s, Getting the instance resource from lister", err.Error())
		}
		log.Println("inside create state")
		log.Println(ingress)

	case "delete":

	case "update":

	}

	return true
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

func (c *Controller) handleIngressUpdate(old interface{}, obj interface{}) {
	log.Println("Updating ingresshandler is called")
	c.state = "update"
	c.wq.Add(obj)
}
