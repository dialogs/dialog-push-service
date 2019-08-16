package metric

import "github.com/prometheus/client_golang/prometheus"

type Peer struct {
	pushRecv prometheus.Counter
}

func (p *Peer) Inc() {
	p.pushRecv.Inc()
}
