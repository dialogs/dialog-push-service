package main

import "github.com/prometheus/client_golang/prometheus"

type metricsCollector struct {
	success *prometheus.CounterVec
	fails   *prometheus.CounterVec
	pushes  *prometheus.CounterVec
	io      *prometheus.HistogramVec
}

type providerMetrics struct {
	success prometheus.Counter
	fails   prometheus.Counter
	pushes  prometheus.Counter
	io      prometheus.Histogram
}

func (m *metricsCollector) getMetricsForProvider(kind, projectId string) (pm *providerMetrics, err error) {
	pm = &providerMetrics{}
	pm.fails, err = m.fails.GetMetricWith(prometheus.Labels{"kind": kind, "projectId": projectId})
	if err != nil {
		return
	}
	pm.success, err = m.success.GetMetricWith(prometheus.Labels{"kind": kind, "projectId": projectId})
	if err != nil {
		return
	}
	pm.pushes, err = m.pushes.GetMetricWith(prometheus.Labels{"kind": kind, "projectId": projectId})
	if err != nil {
		return
	}
	pm.io, err = m.io.GetMetricWith(prometheus.Labels{"kind": kind})
	return
}

func newMetricsCollector() *metricsCollector {
	metrics := &metricsCollector{}
	metrics.success = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "push",
		Name:      "processed_tasks",
		Help:      "Tasks processed by worker"},
		[]string{"kind", "projectId"})
	metrics.fails = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "push",
		Name:      "failed_tasks",
		Help:      "Failed tasks"},
		[]string{"kind", "projectId"})
	metrics.pushes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "push",
		Name:      "pushes_sent",
		Help:      "Pushes sent (w/o result checK)"},
		[]string{"kind", "projectId"})
	metrics.io = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "push",
		Name:      "io",
		Help:      "Time spent in I/O with service provider (in nanoseconds)"},
		[]string{"kind"})
	prometheus.MustRegister(metrics.success, metrics.fails, metrics.pushes, metrics.io)
	return metrics
}
