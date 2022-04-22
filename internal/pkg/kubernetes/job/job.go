package job

import (
	"context"

	batchV1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	jobsV1 "k8s.io/client-go/informers/batch/v1"
	"k8s.io/client-go/kubernetes"
)

type (
	//Job job interface
	Job interface {

		//create job
		Create(ctx context.Context, namespace string, job *batchV1.Job) error

		//delete job
		Delete(ctx context.Context, namespace string, name string) error

		//list jobs
		List(ctx context.Context, namespace string, selector labels.Selector) ([]*batchV1.Job, error)

		//get jobs
		Get(ctx context.Context, namespace string, name string) (*batchV1.Job, error)

		//watch event
		WatchEvent(ctx context.Context, handler JobHandlerFuncs)
	}

	//job object
	jobImpl struct {
		clientset *kubernetes.Clientset
		informer  jobsV1.JobInformer
		factory   informers.SharedInformerFactory
	}

	// JobHandlerFuncs
	JobHandlerFuncs struct {
		AddFunc    func(obj *batchV1.Job)
		UpdateFunc func(oldObj, newObj *batchV1.Job)
		DeleteFunc func(obj *batchV1.Job)
	}
)

//new job
func NewJob(clientset *kubernetes.Clientset, factory informers.SharedInformerFactory) Job {
	//new job
	j := &jobImpl{
		clientset: clientset,
		informer:  factory.Batch().V1().Jobs(),
		factory:   factory,
	}
	return j
}

//create job
func (j *jobImpl) Create(ctx context.Context, namespace string, job *batchV1.Job) error {
	_, err := j.clientset.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

//delete job
func (j *jobImpl) Delete(ctx context.Context, namespace string, name string) error {
	err := j.clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

//list jobs
func (j *jobImpl) List(ctx context.Context, namespace string, selector labels.Selector) (list []*batchV1.Job, err error) {
	list, err = j.informer.Lister().Jobs(namespace).List(selector)
	if err != nil || len(list) <= 0 {
		jobs, err := j.clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return nil, err
		}
		for i := range jobs.Items {
			list = append(list, &jobs.Items[i])
		}
	}
	return list, nil
}

//get jobs
func (j *jobImpl) Get(ctx context.Context, namespace string, name string) (*batchV1.Job, error) {
	job, err := j.informer.Lister().Jobs(namespace).Get(name)
	if err != nil {
		job, err = j.clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
	}
	return job, nil
}

//watch event
func (j *jobImpl) WatchEvent(ctx context.Context, handler JobHandlerFuncs) {
	//add event handler
	j.informer.Informer().AddEventHandler(handler)

	//start
	j.factory.Start(ctx.Done())

	//wait sync
	j.factory.WaitForCacheSync(ctx.Done())
}

// OnAdd calls AddFunc if it's not nil.
func (j JobHandlerFuncs) OnAdd(obj interface{}) {
	if j.AddFunc != nil {
		if event, ok := obj.(*batchV1.Job); ok {
			j.AddFunc(event)
		}
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (j JobHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if j.UpdateFunc != nil {
		old, ok := oldObj.(*batchV1.Job)
		if !ok {
			return
		}
		new, ok := newObj.(*batchV1.Job)
		if !ok {
			return
		}
		j.UpdateFunc(old, new)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (j JobHandlerFuncs) OnDelete(obj interface{}) {
	if j.DeleteFunc != nil {
		if event, ok := obj.(*batchV1.Job); ok {
			j.DeleteFunc(event)
		}
	}
}
