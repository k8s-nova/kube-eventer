package main

import (
	"flag"
	"github.com/k8s-nova/kube-eventer/pkg/collector"
	"github.com/k8s-nova/kube-eventer/pkg/util"
	"github.com/k8s-nova/kube-eventer/pkg/worker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	"log"
	"net/http"
)

var (
	addr       = flag.String("listen", ":9999", "The address to listen on for http requests.")
	kubeConfig = flag.String("kubeconfig", "", "The kubeconfig of the k8s")
)

func main() {
	klog.InitFlags(nil)
	defer klog.Flush()
	flag.Parse()

	stopCh := util.SetupSignalHandler()
	w := worker.NewWorker(*kubeConfig, stopCh)
	go w.Run()

	exporter := collector.NewCollector(&w)
	prometheus.MustRegister(exporter)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
