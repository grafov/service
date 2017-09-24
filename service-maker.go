package service

import (
	"sync"
	"time"
)

var serviceProviders providers

type providers struct {
	sync.RWMutex
	m map[string]*control
}

func init() {
	serviceProviders.m = make(map[string]*control)
}

// Provide signals what something became as a service.
func Provide(name string) *control {
	serviceProviders.Lock()
	c, ok := serviceProviders.m[name]
	if !ok {
		c = &control{
			Name:          name,
			NeedRestart:   make(chan struct{}, 1),
			Dependendents: make(map[string]bool)}
		serviceProviders.m[name] = c
	}
	serviceProviders.Unlock()
	return c
}

// Get returns the service if it started else it returns nil. Use
// WaitFor() if you wish service instance with guarantee.
func Get(name string) interface{} {
	serviceProviders.RLock()
	p := serviceProviders.m[name].Service
	serviceProviders.RUnlock()
	return p
}

// List returns a list of currently ready to use services.
func List() map[string][]string {
	var (
		list = make(map[string][]string)
	)
	serviceProviders.RLock()
	for name, control := range serviceProviders.m {
		if control.IsReady {
			deps := []string{}
			for name := range control.Dependendents {
				deps = append(deps, name)
			}
			list[name] = deps
		}
	}
	serviceProviders.RUnlock()
	return list
}

// Fail signals the service that it is failed with anything
// critical. The service should listen on control.Failed() for these
// signals and handle them with the service restart. After restart
// service should Put() its instance to control again.
func Fail(name string) {
	serviceProviders.RLock()
	c, ok := serviceProviders.m[name]
	serviceProviders.RUnlock()
	if ok {
		c.RLock()
		if !c.IsReady {
			c.RUnlock()
			return
		}
		c.NeedRestart <- struct{}{}
		c.IsReady = false
		for key := range c.Dependendents {
			if name == key {
				continue // cyclic dependency detected
			}
			go func(name string) {
				Fail(name)
			}(key)
		}
		c.RUnlock()
	}
}

type control struct {
	sync.RWMutex
	Name          string
	IsReady       bool
	Service       interface{}
	NeedRestart   chan struct{}
	Dependendents map[string]bool // XXX сюда лучше каналы на рестарт сервисов
}

// WaitFor waits for a service and returns it instance.
func (c *control) WaitFor(name string) interface{} {
	var (
		p       *control
		ok, set bool
	)
	for {
		serviceProviders.RLock()
		p, ok = serviceProviders.m[name]
		serviceProviders.RUnlock()
		if ok {
			if !set {
				p.Lock()
				p.Dependendents[c.Name] = true
				set = true
				p.Unlock()
			}
			p.RLock()
			if p.IsReady {
				s := p.Service
				p.RUnlock()
				return s
			}
			p.RUnlock()
		}
		time.Sleep(10 * time.Millisecond) // XXX alternative with channels
	}
}

// Ready puts the service instance into control structure and says the
// service is ready to serve.
func (c *control) Ready(service interface{}) {
	c.Lock()
	c.Service = service
	c.IsReady = true
	c.Unlock()
}

// Failed notify the service that it is failed and should be
// restarted. The service should listen on control.Failed() for these
// signals and handle them with the service restart. After restart
// service should Put() its instance to control again.
func (c *control) Failed() chan struct{} { // XXX rename
	return c.NeedRestart
}
