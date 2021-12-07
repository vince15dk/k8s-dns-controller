package main

import (
	"flag"
	"fmt"
	"github.com/vince15dk/k8s-operator-ingress/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"time"
)

func main(){
	var kubeconfig *string
	kubeconfig = flag.String("kubeconfig", "/Users/nhn/.kube/config", "location to your kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil{
		// handle error
		fmt.Printf("error %s, building config from flags\n", err.Error())
		config, err = rest.InClusterConfig()
		if err != nil{
			fmt.Printf("error %s, getting inclusterconfig", err.Error())
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil{
		log.Printf("getting std client %s\n", err.Error())
	}

	// period of re-sync by calling UpdateFunc of the event handler
	informer := informers.NewSharedInformerFactory(clientSet, 5*time.Minute)

	ch := make(chan struct{})
	c := controller.NewController(clientSet, informer.Extensions().V1beta1().Ingresses())

	// start initializes all requested informers
	informer.Start(ch)

	if err := c.Run(ch); err != nil{
		log.Printf("error running controller %s\n", err.Error())
	}
}
