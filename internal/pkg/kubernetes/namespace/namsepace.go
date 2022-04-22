package namespace

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	namespacesV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	//namespace interface
	Namespace interface {

		//create namespace
		Create(ctx context.Context, namespace string) error

		//delete namespace
		Delete(ctx context.Context, name string) error

		//list namespaces
		List(ctx context.Context, selector labels.Selector) ([]*coreV1.Namespace, error)

		//get namespaces
		Get(ctx context.Context, name string) (*coreV1.Namespace, error)

		//watch event handler
		WatchEvent(ctx context.Context, handler NamespaceHandlerFuncs)
	}

	//namespace object
	namespaceImpl struct {
		clientset *kubernetes.Clientset
		informer  namespacesV1.NamespaceInformer
		factory   informers.SharedInformerFactory
	}

	// NamespaceHandlerFuncs
	NamespaceHandlerFuncs struct {
		AddFunc    func(obj *coreV1.Namespace)
		UpdateFunc func(oldObj, newObj *coreV1.Namespace)
		DeleteFunc func(obj *coreV1.Namespace)
	}
)

//new namespace
func NewNamespace(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) Namespace {
	//new namespace
	p := &namespaceImpl{
		clientset: clientset,
		informer:  factory.Core().V1().Namespaces(),
		factory:   factory,
	}
	return p
}

//create namespace
func (p *namespaceImpl) Create(ctx context.Context, namespace string) error {
	ns := &coreV1.Namespace{}
	ns.Name = namespace
	_, err := p.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//delete namespace
func (p *namespaceImpl) Delete(ctx context.Context, name string) error {
	err := p.clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list namespaces
func (p *namespaceImpl) List(ctx context.Context, selector labels.Selector) (list []*coreV1.Namespace, err error) {
	list, err = p.informer.Lister().List(selector)
	if err != nil || len(list) <= 0 {
		namespaces, err := p.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		for i := range namespaces.Items {
			list = append(list, &namespaces.Items[i])
		}
	}
	return list, nil
}

//get namespaces
func (p *namespaceImpl) Get(ctx context.Context, name string) (*coreV1.Namespace, error) {
	namespace, err := p.informer.Lister().Get(name)
	if err != nil {
		namespace, err = p.clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return namespace, nil
}

//watch event
func (p *namespaceImpl) WatchEvent(ctx context.Context, handler NamespaceHandlerFuncs) {
	//add event handler
	p.informer.Informer().AddEventHandler(handler)

	//start
	p.factory.Start(ctx.Done())

	//wait sync
	p.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (j NamespaceHandlerFuncs) OnAdd(obj interface{}) {
	if j.AddFunc != nil {
		if event, ok := obj.(*coreV1.Namespace); ok {
			j.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (j NamespaceHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if j.UpdateFunc != nil {
		old, ok := oldObj.(*coreV1.Namespace)
		if !ok {
			return
		}
		new, ok := newObj.(*coreV1.Namespace)
		if !ok {
			return
		}
		j.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (j NamespaceHandlerFuncs) OnDelete(obj interface{}) {
	if j.DeleteFunc != nil {
		if event, ok := obj.(*coreV1.Namespace); ok {
			j.DeleteFunc(event)
		}
	}
}
