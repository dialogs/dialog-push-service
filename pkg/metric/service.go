package metric

import "github.com/prometheus/client_golang/prometheus"

type Service struct {
	success *prometheus.CounterVec
	fails   *prometheus.CounterVec
	io      *prometheus.HistogramVec

	pushesRecv *prometheus.CounterVec
}

func New() *Service {

	m := &Service{
		success: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "push",
			Name:      "processed_tasks",
			Help:      "Tasks processed by worker"},
			[]string{"kind", "projectId"}),
		fails: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "push",
			Name:      "failed_tasks",
			Help:      "Failed tasks"},
			[]string{"kind", "projectId"}),
		io: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "push",
			Name:      "io",
			Help:      "Time spent in I/O with service provider (in nanoseconds)"},
			[]string{"kind"}),
		pushesRecv: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "push",
			Name:      "pushes_recv",
			Help:      "Pushes recv"},
			[]string{"addr"}),
	}

	for _, c := range []prometheus.Collector{
		m.success,
		m.fails,
		m.io,
		m.pushesRecv,
	} {
		if err := prometheus.Register(c); err != nil {
			switch err.(type) {
			case prometheus.AlreadyRegisteredError:
				break
			default:
				panic(err)
			}
		}
	}

	return m
}

func (m *Service) GetProviderMetrics(kind, projectId string) (*Provider, error) {

	var err error

	p := &Provider{}
	p.fails, err = m.fails.GetMetricWith(prometheus.Labels{"kind": kind, "projectId": projectId})
	if err != nil {
		return nil, err
	}

	p.success, err = m.success.GetMetricWith(prometheus.Labels{"kind": kind, "projectId": projectId})
	if err != nil {
		return nil, err
	}

	p.io, err = m.io.GetMetricWith(prometheus.Labels{"kind": kind})
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (m *Service) GetPeerMetrics(addr string) (*Peer, error) {

	pushRecv, err := m.pushesRecv.GetMetricWith(prometheus.Labels{"addr": addr})
	if err != nil {
		return nil, err
	}

	return &Peer{pushRecv: pushRecv}, nil
}
