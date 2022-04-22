package pod

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	podsV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	//pod interface
	Pod interface {

		//create pod
		Create(ctx context.Context, namespace string, pod *coreV1.Pod) error

		//delete pod
		Delete(ctx context.Context, namespace string, name string) error

		//list pods
		List(ctx context.Context, namespace string, selector labels.Selector) ([]*coreV1.Pod, error)

		//get pods
		Get(ctx context.Context, namespace string, name string) (*coreV1.Pod, error)

		//watch event handler
		WatchEvent(ctx context.Context, handler PodHandlerFuncs)
	}

	//pod object
	podImpl struct {
		clientset *kubernetes.Clientset
		informer  podsV1.PodInformer
		factory   informers.SharedInformerFactory
	}

	// PodHandlerFuncs
	PodHandlerFuncs struct {
		AddFunc    func(obj *coreV1.Pod)
		UpdateFunc func(oldObj, newObj *coreV1.Pod)
		DeleteFunc func(obj *coreV1.Pod)
	}
)

//new pod
func NewPod(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) Pod {
	//new pod
	p := &podImpl{
		clientset: clientset,
		informer:  factory.Core().V1().Pods(),
		factory:   factory,
	}
	return p
}

//create pod
func (p *podImpl) Create(ctx context.Context, namespace string, pod *coreV1.Pod) error {
	_, err := p.clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//delete pod
func (p *podImpl) Delete(ctx context.Context, namespace string, name string) error {
	err := p.clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list pods
func (p *podImpl) List(ctx context.Context, namespace string, selector labels.Selector) (list []*coreV1.Pod, err error) {
	list, err = p.informer.Lister().Pods(namespace).List(selector)
	if err != nil || len(list) <= 0 {
		pods, err := p.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		for i := range pods.Items {
			list = append(list, &pods.Items[i])
		}
	}
	return list, nil
}

//get pods
func (p *podImpl) Get(ctx context.Context, namespace string, name string) (*coreV1.Pod, error) {
	pod, err := p.informer.Lister().Pods(namespace).Get(name)
	if err != nil {
		pod, err = p.clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return pod, nil
}

//watch event
func (p *podImpl) WatchEvent(ctx context.Context, handler PodHandlerFuncs) {
	//add event handler
	p.informer.Informer().AddEventHandler(handler)

	//start
	p.factory.Start(ctx.Done())

	//wait sync
	p.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (j PodHandlerFuncs) OnAdd(obj interface{}) {
	if j.AddFunc != nil {
		if event, ok := obj.(*coreV1.Pod); ok {
			j.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (j PodHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if j.UpdateFunc != nil {
		old, ok := oldObj.(*coreV1.Pod)
		if !ok {
			return
		}
		new, ok := newObj.(*coreV1.Pod)
		if !ok {
			return
		}
		j.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (j PodHandlerFuncs) OnDelete(obj interface{}) {
	if j.DeleteFunc != nil {
		if event, ok := obj.(*coreV1.Pod); ok {
			j.DeleteFunc(event)
		}
	}
}
