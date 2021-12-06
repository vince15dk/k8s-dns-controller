package code

/*
func (c *Controller) waitForIngressLBPoll(namespace, name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	lb := ""
	err := poll.Wait(ctx, func(ctx context.Context) (bool, error) {
		ingress, err := c.client.ExtensionsV1beta1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			log.Printf("err %s\n", err.Error())
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
*/

