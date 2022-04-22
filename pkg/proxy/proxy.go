package proxy

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/kubegames/kubegames-proxy/internal/pkg/kubernetes/pod"
	"github.com/kubegames/kubegames-proxy/internal/pkg/kubernetes/service"
	"github.com/kubegames/kubegames-proxy/internal/pkg/log"
	"github.com/kubegames/kubegames-proxy/pkg/route"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

const (
	//labels proxy
	LabelsProxy = "proxy"
	//format time usage
	timeFormat = "2006-01-02 15:04:05"
)

type (
	ProxyApp interface {
		//start
		Start(ctx context.Context) error
	}

	proxyAppImp struct {
		route   *route.Route
		port    string
		pod     pod.Pod
		service service.Service
	}
)

//new app object impl
func NewProxyApp(port string, clientset *kubernetes.Clientset) ProxyApp {
	//new share informer factory
	factory := informers.NewSharedInformerFactory(clientset, 0)

	//new game impl
	app := &proxyAppImp{
		pod:     pod.NewPod(clientset, factory),
		service: service.NewService(clientset, factory),
		port:    port,
		route:   route.NewRoute(),
	}
	return app
}

//start
func (app *proxyAppImp) Start(ctx context.Context) error {
	//pod event
	app.pod.WatchEvent(ctx, pod.PodHandlerFuncs{
		AddFunc: func(obj *v1.Pod) {
			app.AddPodRoute(obj)
		},
		UpdateFunc: func(oldObj, newObj *v1.Pod) {
			if oldObj.ResourceVersion != newObj.ResourceVersion {
				obj, err := app.pod.Get(ctx, newObj.Namespace, newObj.Name)
				if err != nil {
					return
				}
				app.AddPodRoute(obj)
			}
		},
		DeleteFunc: func(obj *v1.Pod) {
			app.DeletePodRoute(obj)
		},
	})

	//service event
	app.service.WatchEvent(ctx, service.ServiceHandlerFuncs{
		AddFunc: func(obj *v1.Service) {
			app.AddServiceRoute(obj)
		},
		UpdateFunc: func(oldObj, newObj *v1.Service) {
			if oldObj.ResourceVersion != newObj.ResourceVersion {
				obj, err := app.service.Get(ctx, newObj.Namespace, newObj.Name)
				if err != nil {
					return
				}
				app.AddServiceRoute(obj)
			}
		},
		DeleteFunc: func(obj *v1.Service) {
			app.DeleteServiceRoute(obj)
		},
	})

	go app.Http()

	log.Infof("proxy app start %s", app.port)
	<-ctx.Done()
	return nil
}

func (app *proxyAppImp) AddServiceRoute(obj *v1.Service) {
	proxy, ok := obj.Annotations[LabelsProxy]
	if !ok {
		return
	}

	//rule
	rules, err := route.Unmarshal(proxy)
	if err != nil {
		log.Errorf("rule unmarshal err %s", err.Error())
		return
	}

	//proxy patten
	if rules.ProxyPattern != route.Service {
		log.Warnln("rule ProxyPattern != Service")
		return
	}

	//register route
	for _, rule := range rules.Items {
		//get proxy ip
		proxyIp := fmt.Sprintf("http://%s:%d", obj.Spec.ClusterIP, rule.Port)

		//add
		if rule.Method == route.Any {
			app.route.Add(route.GET, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.POST, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.DELETE, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.PUT, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.PATCH, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.HEAD, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.TRACE, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.OPTIONS, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.CONNECT, rule.AgentUrl, rule.ProxyUrl, proxyIp)
		} else {
			app.route.Add(rule.Method, rule.AgentUrl, rule.ProxyUrl, proxyIp)
		}
	}
}

func (app *proxyAppImp) DeleteServiceRoute(obj *v1.Service) {
	proxy, ok := obj.Annotations[LabelsProxy]
	if !ok {
		return
	}

	//port
	if len(obj.Spec.Ports) <= 0 {
		return
	}

	//rule
	rules, err := route.Unmarshal(proxy)
	if err != nil {
		log.Errorf("rule unmarshal err %s", err.Error())
		return
	}

	//proxy patten
	if rules.ProxyPattern != route.Service {
		log.Warnln("rule ProxyPattern != Service")
		return
	}

	//delete route
	for _, rule := range rules.Items {
		if rule.Method == route.Any {
			app.route.Delete(route.GET, rule.AgentUrl)
			app.route.Delete(route.POST, rule.AgentUrl)
			app.route.Delete(route.DELETE, rule.AgentUrl)
			app.route.Delete(route.PUT, rule.AgentUrl)
			app.route.Delete(route.PATCH, rule.AgentUrl)
			app.route.Delete(route.HEAD, rule.AgentUrl)
			app.route.Delete(route.TRACE, rule.AgentUrl)
			app.route.Delete(route.OPTIONS, rule.AgentUrl)
			app.route.Delete(route.CONNECT, rule.AgentUrl)
		} else {
			app.route.Delete(rule.Method, rule.AgentUrl)
		}
	}
}

func (app *proxyAppImp) AddPodRoute(obj *v1.Pod) {
	proxy, ok := obj.Annotations[LabelsProxy]
	if !ok {
		return
	}

	//check pod is running
	if obj.Status.Phase != v1.PodRunning {
		return
	}

	//rule
	rules, err := route.Unmarshal(proxy)
	if err != nil {
		log.Errorf("rule unmarshal err %s", err.Error())
		return
	}

	//proxy patten
	if rules.ProxyPattern != route.Pod {
		log.Warnln("rule ProxyPattern != Pod")
		return
	}

	//register route
	for _, rule := range rules.Items {
		//get proxy ip
		proxyIp := fmt.Sprintf("http://%s:%d", obj.Status.PodIP, rule.Port)

		if rule.Method == route.Any {
			app.route.Add(route.GET, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.POST, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.DELETE, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.PUT, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.PATCH, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.HEAD, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.TRACE, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.OPTIONS, rule.AgentUrl, rule.ProxyUrl, proxyIp)
			app.route.Add(route.CONNECT, rule.AgentUrl, rule.ProxyUrl, proxyIp)
		} else {
			app.route.Add(rule.Method, rule.AgentUrl, rule.ProxyUrl, proxyIp)
		}
	}
}

func (app *proxyAppImp) DeletePodRoute(obj *v1.Pod) {
	proxy, ok := obj.Annotations[LabelsProxy]
	if !ok {
		return
	}

	//check pod is running
	if obj.Status.Phase != v1.PodRunning {
		return
	}

	if len(obj.Spec.Containers) <= 0 {
		return
	}

	if len(obj.Spec.Containers[0].Ports) <= 0 {
		return
	}

	//rule
	rules, err := route.Unmarshal(proxy)
	if err != nil {
		log.Errorf("rule unmarshal err %s", err.Error())
		return
	}

	//proxy patten
	if rules.ProxyPattern != route.Pod {
		log.Warnln("rule ProxyPattern != Pod")
		return
	}

	//delete route
	for _, rule := range rules.Items {
		if rule.Method == route.Any {
			app.route.Delete(route.GET, rule.AgentUrl)
			app.route.Delete(route.POST, rule.AgentUrl)
			app.route.Delete(route.DELETE, rule.AgentUrl)
			app.route.Delete(route.PUT, rule.AgentUrl)
			app.route.Delete(route.PATCH, rule.AgentUrl)
			app.route.Delete(route.HEAD, rule.AgentUrl)
			app.route.Delete(route.TRACE, rule.AgentUrl)
			app.route.Delete(route.OPTIONS, rule.AgentUrl)
			app.route.Delete(route.CONNECT, rule.AgentUrl)
		} else {
			app.route.Delete(rule.Method, rule.AgentUrl)
		}
	}
}

//http proxy run
func (app *proxyAppImp) Http() error {
	http.HandleFunc("/", app.Proxy)
	err := http.ListenAndServe(app.port, nil)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

//proxy
func (app *proxyAppImp) Proxy(w http.ResponseWriter, r *http.Request) {
	//get method
	method := route.Method(strings.ToUpper(r.Method))

	//find route
	log.Tracef("find url %s", r.URL.Path)
	proxyIp, proxyPath, ok := app.route.Find(method, r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "not found")
		return
	}

	if proxyPath == "/" {
		proxyPath = ""
	}

	//create proxy
	remote, err := url.Parse(proxyIp)
	if err != nil {
		log.Errorln(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "bad request")
		return
	}

	//set path
	r.URL.Path = proxyPath

	//proxy
	log.Tracef("proxy %s ===> %s", proxyIp, proxyPath)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
	return
}
