package main

import (
	"flag"
	"github.com/vince15dk/k8s-operator-ingress/pkg/controller"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"
)

func main(){
	var kubeconfig *string
	if home := homedir.HomeDir(); home != ""{
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}else{
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil{
		log.Printf("Building ocnfig from flags, %s", err.Error())
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
