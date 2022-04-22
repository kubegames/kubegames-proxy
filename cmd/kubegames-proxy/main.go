package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/kubegames/kubegames-proxy/pkg/proxy"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

//config cmd
var (
	help       bool
	port       int64
	cfg        string
	kubeconfig string
)

func init() {
	flag.Int64Var(&port, "p", 8080, "http server port")
	flag.Int64Var(&port, "port", 8080, "http server port")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) kubeconfig absolute path to the file")
		flag.StringVar(&kubeconfig, "k", filepath.Join(home, ".kube", "config"), "(optional) kubeconfig absolute path to the file")
	} else {
		flag.StringVar(&kubeconfig, "kubeconfig", "", "(optional) kubeconfig absolute path to the file")
		flag.StringVar(&kubeconfig, "k", "", "(optional) kubeconfig absolute path to the file")
	}
	flag.Parse()
}

func main() {
	//flag
	if help {
		flag.Usage()
		return
	}

	//nee k8s client
	var config *rest.Config
	var err error

	if len(kubeconfig) > 0 {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			panic(err.Error())
		}
	} else {
		if config, err = rest.InClusterConfig(); err != nil {
			panic(err.Error())
		}
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//new with cancel context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//listen signal
	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

	//run server
	go func() {
		//start
		app := proxy.NewProxyApp(fmt.Sprintf(":%d", port), kubeClient)
		if err := app.Start(ctx); err != nil {
			panic(err.Error())
		}
	}()

	//wait
	signal := <-c
	cancel()
	if signal != nil {
		log.Fatalf("received signal %s", signal)
	}
}
