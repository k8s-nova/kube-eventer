package worker

import (
	v1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type Worker struct {
	mu sync.Mutex
	events *[]Event
	kubeConfig string
	stopCh <-chan struct{}
}

func NewWorker(kubeConfig string, stopCh <-chan struct{}) Worker {
	w := Worker{
		kubeConfig: kubeConfig,
		stopCh: stopCh,
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
	eventsCurrent, err := kubeInformerFactory.Events().V1().Events().Lister().List(labels.NewSelector())
	if err != nil {
		klog.Error(err)
	}
	if len(eventsCurrent) > 0 {
		w.eventInitHandle(eventsCurrent)
	}
	eventInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: w.eventAddHandle,
	})
}

func (w *Worker) eventInitHandle(es []*v1.Event)  {
	// Todo: init, list all k8s events ---> Worker.events
	for _, event := range es {
		var eventTmp Event
		eventTmp.Type = event.Type
		eventTmp.Kind = event.Regarding.Kind
		eventTmp.Name = event.Regarding.Name
		eventTmp.Message = event.Note
		eventTmp.Host = event.DeprecatedSource.Host
		eventTmp.Namespace = event.Namespace
		eventTmp.count = event.DeprecatedCount
		eventTmp.Reason = event.Reason
		eventTmp.Source = event.DeprecatedSource.Component
		eventTmp.Timestamp = event.DeprecatedLastTimestamp.Time
	}
}

func (w *Worker) eventAddHandle(obj interface{})  {
	// Todo: watch add-event, and update events ---> Worker events
}