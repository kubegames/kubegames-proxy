package configmap

import (
	"context"

	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	configmapV1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	//configmap interface
	ConfigMap interface {

		//create configmap
		Create(ctx context.Context, namespace string, cm *coreV1.ConfigMap) error

		//update configmap
		Update(ctx context.Context, namespace string, cm *coreV1.ConfigMap) error

		//delete configmap
		Delete(ctx context.Context, namespace string, name string) error

		//list configmap
		List(ctx context.Context, namespace string, selector labels.Selector) ([]*coreV1.ConfigMap, error)

		//get configmap
		Get(ctx context.Context, namespace string, name string) (*coreV1.ConfigMap, error)

		//watch event
		WatchEvent(ctx context.Context, handler ConfigMapHandlerFuncs)
	}

	//configmap impl
	configMapImpl struct {
		clientset *kubernetes.Clientset
		informer  configmapV1.ConfigMapInformer
		factory   informers.SharedInformerFactory
	}

	// ConfigMapHandlerFuncs
	ConfigMapHandlerFuncs struct {
		AddFunc    func(obj *coreV1.ConfigMap)
		UpdateFunc func(oldObj, newObj *coreV1.ConfigMap)
		DeleteFunc func(obj *coreV1.ConfigMap)
	}
)

//new configmap
func NewConfigMap(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) ConfigMap {
	//new config map
	j := &configMapImpl{
		clientset: clientset,
		informer:  factory.Core().V1().ConfigMaps(),
		factory:   factory,
	}
	return j
}

//create configmap
func (j *configMapImpl) Create(ctx context.Context, namespace string, cm *coreV1.ConfigMap) error {
	_, err := j.clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//delete configmap
func (j *configMapImpl) Delete(ctx context.Context, namespace string, name string) error {
	err := j.clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//update configmap
func (j *configMapImpl) Update(ctx context.Context, namespace string, cm *coreV1.ConfigMap) error {
	_, err := j.clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list configmap
func (j *configMapImpl) List(ctx context.Context, namespace string, selector labels.Selector) (list []*coreV1.ConfigMap, err error) {
	list, err = j.informer.Lister().ConfigMaps(namespace).List(selector)
	if err != nil || len(list) <= 0 {
		jobs, err := j.clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		for i := range jobs.Items {
			list = append(list, &jobs.Items[i])
		}
	}
	return list, nil
}

//get configmap
func (j *configMapImpl) Get(ctx context.Context, namespace string, name string) (*coreV1.ConfigMap, error) {
	job, err := j.informer.Lister().ConfigMaps(namespace).Get(name)
	if err != nil {
		job, err = j.clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return job, nil
}

//watch event
func (j *configMapImpl) WatchEvent(ctx context.Context, handler ConfigMapHandlerFuncs) {
	//add event handler
	j.informer.Informer().AddEventHandler(handler)

	//start
	j.factory.Start(ctx.Done())

	//wait sync
	j.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (j ConfigMapHandlerFuncs) OnAdd(obj interface{}) {
	if j.AddFunc != nil {
		if event, ok := obj.(*coreV1.ConfigMap); ok {
			j.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (j ConfigMapHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if j.UpdateFunc != nil {
		old, ok := oldObj.(*coreV1.ConfigMap)
		if !ok {
			return
		}
		new, ok := newObj.(*coreV1.ConfigMap)
		if !ok {
			return
		}
		j.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (j ConfigMapHandlerFuncs) OnDelete(obj interface{}) {
	if j.DeleteFunc != nil {
		if event, ok := obj.(*coreV1.ConfigMap); ok {
			j.DeleteFunc(event)
		}
	}
}
