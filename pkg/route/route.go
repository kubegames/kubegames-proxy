package route

import (
	"strings"
	"sync"

	"github.com/kubegames/kubegames-proxy/internal/pkg/log"
)

type (
	Route struct {
		Node map[Method]*Node
		lock sync.RWMutex
	}

	Node struct {
		Route    map[string]*Node
		ProxyUrl string
		ProxyIp  string
	}
)

//new route
func NewRoute() *Route {
	return &Route{
		Node: make(map[Method]*Node),
	}
}

func (n *Node) Add(path string) *Node {
	node, ok := n.Route[path]
	if !ok {
		//new node
		node = new(Node)
		node.Route = make(map[string]*Node)
	}

	//insert node
	n.Route[path] = node
	return node
}

func (n *Node) Find(path string) (*Node, bool) {
	node, ok := n.Route[path]
	if !ok {
		return nil, false
	}
	return node, true
}

func (n *Node) Delete(path string) (*Node, bool) {
	node, ok := n.Route[path]
	if !ok {
		return nil, false
	}
	delete(n.Route, path)
	return node, true
}

//add route
func (r *Route) Add(method Method, agentUrl string, proxyUrl string, proxyIp string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	node, ok := r.Node[method]
	if !ok {
		//new node
		node = new(Node)
		node.Route = make(map[string]*Node)
	}
	r.Node[method] = node

	paths := strings.Split(agentUrl, "/")
	for _, path := range paths {
		if len(path) <= 0 {
			continue
		}
		node = node.Add(path)
	}

	//set
	node.ProxyUrl = proxyUrl
	node.ProxyIp = proxyIp
	log.Tracef("add proxy %s %s ===> %s%s", method, agentUrl, node.ProxyIp, node.ProxyUrl)
}

//delete
func (r *Route) Delete(method Method, url string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	node, ok := r.Node[method]
	if !ok {
		return
	}

	paths := strings.Split(url, "/")
	for _, path := range paths {
		if len(path) <= 0 {
			continue
		}
		n, ok := node.Delete(path)
		if !ok {
			break
		}
		node = n
	}
	log.Tracef("delete proxy %s %s", method, url)
}

//find
func (r *Route) Find(method Method, url string) (string, string, bool) {
	r.lock.Lock()
	defer r.lock.Unlock()

	node, ok := r.Node[method]
	if !ok {
		return "", "", false
	}

	paths := strings.Split(url, "/")
	route := ""
	for index, path := range paths {
		if len(path) <= 0 {
			continue
		}
		n, ok := node.Find(path)
		if !ok {
			proxy := strings.Split(node.ProxyUrl, "/")
			proxy = append(proxy, paths[index:]...)
			route = strings.Join(proxy, "/")
			break
		}
		node = n
	}

	//check
	if len(node.ProxyIp) <= 0 {
		return "", "", false
	}

	//return
	return node.ProxyIp, route, true
}
