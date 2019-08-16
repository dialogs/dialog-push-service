package metric

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	success prometheus.Counter
	fails   prometheus.Counter
	io      prometheus.Observer
}

func (p *Provider) SuccessInc() {
	p.success.Inc()
}

func (p *Provider) FailsInc() {
	p.fails.Inc()
}

func (p *Provider) NewIOTimer() (cancel func()) {
	timer := prometheus.NewTimer(p.io)
	cancel = func() {
		timer.ObserveDuration()
	}
	return
}
