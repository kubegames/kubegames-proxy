package event

import (
	"context"

	v1 "k8s.io/api/events/v1beta1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/informers"
	eventsV1 "k8s.io/client-go/informers/events/v1beta1"
	"k8s.io/client-go/kubernetes"
)

type (
	//event interface
	Event interface {

		//delete event
		Delete(ctx context.Context, namespace string, name string) error

		//list event
		List(ctx context.Context, namespace string, fieldselector fields.Selector) ([]*v1.Event, error)

		//get event
		Get(ctx context.Context, namespace string, name string) (event *v1.Event, err error)

		//watch event
		WatchEvent(ctx context.Context, handler EventHandlerFuncs)
	}

	//event
	eventImpl struct {
		clientset *kubernetes.Clientset
		informer  eventsV1.EventInformer
		factory   informers.SharedInformerFactory
		handler   EventHandlerFuncs
	}

	// EventHandlerFuncs
	EventHandlerFuncs struct {
		AddFunc    func(obj *v1.Event)
		UpdateFunc func(oldObj, newObj *v1.Event)
		DeleteFunc func(obj *v1.Event)
	}
)

//new event
func NewEvent(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) Event {
	//new event
	e := &eventImpl{
		informer:  factory.Events().V1beta1().Events(),
		clientset: clientset,
		factory:   factory,
	}
	return e
}

//delete event
func (e *eventImpl) Delete(ctx context.Context, namespace string, name string) error {
	err := e.clientset.EventsV1beta1().Events(namespace).Delete(ctx, name, metaV1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list event
func (e *eventImpl) List(ctx context.Context, namespace string, fieldselector fields.Selector) (list []*v1.Event, err error) {
	events, err := e.clientset.EventsV1beta1().Events(namespace).List(ctx, metaV1.ListOptions{
		FieldSelector: fieldselector.String(),
	})
	if err != nil {
		return nil, err
	}
	for i := range events.Items {
		list = append(list, &events.Items[i])
	}
	return list, nil
}

//get event
func (e *eventImpl) Get(ctx context.Context, namespace string, name string) (event *v1.Event, err error) {
	event, err = e.informer.Lister().Events(namespace).Get(name)
	if err != nil {
		event, err = e.clientset.EventsV1beta1().Events(namespace).Get(ctx, name, metaV1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return event, nil
}

//watch event
func (e *eventImpl) WatchEvent(ctx context.Context, handler EventHandlerFuncs) {
	//add event handler
	e.informer.Informer().AddEventHandler(handler)

	//start
	e.factory.Start(ctx.Done())

	//wait sync
	e.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (r EventHandlerFuncs) OnAdd(obj interface{}) {
	if r.AddFunc != nil {
		if event, ok := obj.(*v1.Event); ok {
			r.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r EventHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if r.UpdateFunc != nil {
		old, ok := oldObj.(*v1.Event)
		if !ok {
			return
		}
		new, ok := newObj.(*v1.Event)
		if !ok {
			return
		}
		r.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (r EventHandlerFuncs) OnDelete(obj interface{}) {
	if r.DeleteFunc != nil {
		if event, ok := obj.(*v1.Event); ok {
			r.DeleteFunc(event)
		}
	}
}
