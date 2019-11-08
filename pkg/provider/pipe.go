package provider

import (
	"context"
	"io"
	"sync"
)

type Pipe struct {
	r         *io.PipeReader
	ctx       context.Context
	ctxCancel func()
	encodeErr error
	mu        sync.RWMutex
	wg        sync.WaitGroup
}

func NewPipe(write func(io.Writer) error) *Pipe {

	ctx, ctxCancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()

	p := &Pipe{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		r:         pr,
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.ctxCancel()

		select {
		case <-p.ctx.Done():
			_ = pw.Close()

		case <-func() <-chan struct{} {
			chWait := make(chan struct{})

			go func() {
				defer close(chWait)
				defer pw.Close()

				err := write(pw)

				select {
				case <-p.ctx.Done():
					if err == io.ErrClosedPipe {
						err = io.EOF // already closed
					}
				default:
					// nothing do
				}

				p.mu.Lock()
				p.encodeErr = err
				p.mu.Unlock()
			}()

			return chWait
		}():
		}

	}()

	return p
}

func (p *Pipe) Read(out []byte) (n int, err error) {

	if intError := p.getErr(); intError != nil {
		return 0, intError
	}

	n, err = p.r.Read(out)
	if err == nil {
		if intError := p.getErr(); intError != nil {
			err = intError
		}
	}

	return
}

func (p *Pipe) Close() error {
	p.ctxCancel()

	p.mu.Lock()
	p.wg.Wait()
	p.mu.Unlock()

	err := p.getErr()
	if err == io.EOF {
		return nil
	}

	return err
}

func (p *Pipe) getErr() (err error) {
	p.mu.RLock()
	err = p.encodeErr
	p.mu.RUnlock()

	return
}
