package service

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	serviceV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	//service interface
	Service interface {

		//create service
		Create(ctx context.Context, namespace string, service *coreV1.Service) error

		//delete service
		Delete(ctx context.Context, namespace string, name string) error

		//list service
		List(ctx context.Context, namespace string, selector labels.Selector) ([]*coreV1.Service, error)

		//get service
		Get(ctx context.Context, namespace string, name string) (*coreV1.Service, error)

		//watch event handler
		WatchEvent(ctx context.Context, handler ServiceHandlerFuncs)
	}

	//service object
	serviceImpl struct {
		clientset *kubernetes.Clientset
		informer  serviceV1.ServiceInformer
		factory   informers.SharedInformerFactory
	}

	// ServiceHandlerFuncs
	ServiceHandlerFuncs struct {
		AddFunc    func(obj *coreV1.Service)
		UpdateFunc func(oldObj, newObj *coreV1.Service)
		DeleteFunc func(obj *coreV1.Service)
	}
)

//new service
func NewService(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) Service {
	//new service
	p := &serviceImpl{
		clientset: clientset,
		informer:  factory.Core().V1().Services(),
		factory:   factory,
	}
	return p
}

//create service
func (p *serviceImpl) Create(ctx context.Context, namespace string, service *coreV1.Service) error {
	_, err := p.clientset.CoreV1().Services(namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//delete service
func (p *serviceImpl) Delete(ctx context.Context, namespace string, name string) error {
	err := p.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list service
func (p *serviceImpl) List(ctx context.Context, namespace string, selector labels.Selector) (list []*coreV1.Service, err error) {
	list, err = p.informer.Lister().Services(namespace).List(selector)
	if err != nil || len(list) <= 0 {
		service, err := p.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		for i := range service.Items {
			list = append(list, &service.Items[i])
		}
	}
	return list, nil
}

//get service
func (p *serviceImpl) Get(ctx context.Context, namespace string, name string) (*coreV1.Service, error) {
	service, err := p.informer.Lister().Services(namespace).Get(name)
	if err != nil {
		service, err = p.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return service, nil
}

//watch event
func (p *serviceImpl) WatchEvent(ctx context.Context, handler ServiceHandlerFuncs) {
	//add event handler
	p.informer.Informer().AddEventHandler(handler)

	//start
	p.factory.Start(ctx.Done())

	//wait sync
	p.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (j ServiceHandlerFuncs) OnAdd(obj interface{}) {
	if j.AddFunc != nil {
		if event, ok := obj.(*coreV1.Service); ok {
			j.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (j ServiceHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if j.UpdateFunc != nil {
		old, ok := oldObj.(*coreV1.Service)
		if !ok {
			return
		}
		new, ok := newObj.(*coreV1.Service)
		if !ok {
			return
		}
		j.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (j ServiceHandlerFuncs) OnDelete(obj interface{}) {
	if j.DeleteFunc != nil {
		if event, ok := obj.(*coreV1.Service); ok {
			j.DeleteFunc(event)
		}
	}
}
