package collector

import (
	"github.com/k8s-nova/kube-eventer/pkg/worker"
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

type collector struct {
	metrics *prometheus.GaugeVec
	events  *[]worker.Event
	mutex   *sync.Mutex
}

func (c *collector) Describe(desc chan<- *prometheus.Desc) {
	c.metrics.Describe(desc)
}

func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	if c.events == nil || len(*c.events) == 0 {
		return
	}
	c.mutex.Lock()
	c.metrics.Reset()
	for _, e := range *c.events {
		label := map[string]string{
			"type": e.Type,
			"kind": e.Kind,
			"name": e.Name,
			"namespace": e.Namespace,
			"timestamp": e.Timestamp.Format("2006-01-02-15:04:05"),
			"message": e.Message,
			"reason": e.Reason,
			"source": e.Source,
			"host": e.Host,
		}
		c.metrics.With(label).Set(float64(e.Count))
	}
	c.metrics.Collect(metrics)
	*c.events = *new([]worker.Event)
	c.mutex.Unlock()
}

func NewCollector(w *worker.Worker) *collector {
	e := collector{
		metrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k8s_event_count",
				Help: "The current number of event's count",
			},
			[]string{"type", "kind", "name", "namespace", "timestamp", "message", "reason", "source", "host"},
		),
		events: &w.Events,
		mutex: w.Mutex,
	}
	return &e
}
