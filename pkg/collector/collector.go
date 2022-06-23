package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"math/rand"
	"strings"
)

type collector struct {
	metrics      *prometheus.GaugeVec
	customLabel  string
}

func (c *collector) Describe(desc chan<- *prometheus.Desc) {
	c.metrics.Describe(desc)
}

func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	if c.customLabel != "" {
		labelMap := strings.Split(c.customLabel, "=")
		label := map[string]string{labelMap[0]: labelMap[1]}
		c.metrics.With(label).Set(float64(rand.Int()))
	} else {
		label := map[string]string{"demo": "xc"}
		c.metrics.With(label).Set(float64(rand.Int()))
	}
	c.metrics.Collect(metrics)
}

func NewCollector(customLabel *string) *collector {
	e := collector{
		metrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "k8s_event_count",
				Help: "The current number of event's count",
			},
			[]string{"demo"},
		) ,
		customLabel: *customLabel,
	}
	return &e
}