package metric

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Provider struct {
	success prometheus.Counter
	fails   prometheus.Counter
	pushes  prometheus.Counter
	io      prometheus.Observer
}

func (p *Provider) SuccessInc() {
	p.success.Inc()
}

func (p *Provider) FailsInc() {
	p.fails.Inc()
}

func (p *Provider) PushesInc(t time.Time) {
	p.io.Observe(float64(time.Since(t).Nanoseconds()))
	p.pushes.Inc()
}
