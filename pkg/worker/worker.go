package worker

import (
	"encoding/json"
	v1 "k8s.io/api/events/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"log"
	"os"
	"sync"
	"time"
)

type Worker struct {
	Mutex      *sync.Mutex
	Events     []Event
	kubeConfig string
	stopCh     <-chan struct{}
}

func NewWorker(kubeConfig string, stopCh <-chan struct{}) Worker {
	w := Worker{
		Mutex:      new(sync.Mutex),
		Events:     *new([]Event),
		kubeConfig: kubeConfig,
		stopCh:     stopCh,
	}
	return w
}

func (w *Worker) Run() {
	config, _ := clientcmd.BuildConfigFromFlags("", w.kubeConfig)
	clientSet, _ := kubernetes.NewForConfig(config)
	kubeInformerFactory := informers.NewSharedInformerFactory(clientSet, time.Second*30)
	eventInformer := kubeInformerFactory.Events().V1().Events().Informer()
	go kubeInformerFactory.Start(w.stopCh)
	if !cache.WaitForCacheSync(w.stopCh, eventInformer.HasSynced) {
		klog.Fatal("Timed out waiting for caches to sync")
	}

	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: w.eventAddHandle,
	})
}

func (w *Worker) eventAddHandle(obj interface{}) {
	// watch add-event, and update events ---> Worker events
	event := obj.(*v1.Event)
	var eventTmp Event
	eventTmp.Type = event.Type
	eventTmp.Kind = event.Regarding.Kind
	eventTmp.Name = event.Regarding.Name
	eventTmp.Message = event.Note
	eventTmp.Host = event.DeprecatedSource.Host
	eventTmp.Namespace = event.Namespace
	eventTmp.Count = event.DeprecatedCount
	eventTmp.Reason = event.Reason
	eventTmp.Source = event.DeprecatedSource.Component
	eventTmp.Timestamp = event.DeprecatedLastTimestamp.Time
	w.Mutex.Lock()
	w.Events = append(w.Events, eventTmp)
	w.Mutex.Unlock()
	klog.V(2).Infof("Add a event to worker.events, len: %d", len(w.Events))

	// Send event to stdout.
	logger := log.New(os.Stdout, "", 0)
	writer := logger.Writer()
	err := json.NewEncoder(writer).Encode(event)
	if err != nil {
		klog.Error("Failed to send event to stdout, err: %s", err.Error())
	}
}
