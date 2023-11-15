package main

import (
  "github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
  concurrentExecutions prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *metrics {
  m := &metrics{
    concurrentExecutions: prometheus.NewGauge(prometheus.GaugeOpts{
      Namespace: "myapp",
      Name: "concurrent_executions",
      Help: "Number of concurrent executions",
    }),
  }

  reg.MustRegister(m.concurrentExecutions)
  return m
}
