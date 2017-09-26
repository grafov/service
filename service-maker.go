package service

import (
	"sync"
	"time"
)

var serviceProviders providers

type providers struct {
	sync.RWMutex
	m map[string]*service
}

func init() {
	serviceProviders.m = make(map[string]*service)
}

// Free to add something useful here like logging or metrics.
var (
	EnterActionHook func(action, name string) = func(action, name string) { return }
	ExitActionHook  func(action, name string) = func(action, name string) { return }
)

// Provide signals what something became as a service.
func Provide(name string) *service {
	EnterActionHook("register", name)
	serviceProviders.Lock()
	c, ok := serviceProviders.m[name]
	if !ok {
		c = &service{
			Name:          name,
			NeedRestart:   make(chan struct{}, 1),
			Dependendents: make(map[string]bool)}
		serviceProviders.m[name] = c
	}
	serviceProviders.Unlock()
	ExitActionHook("register", name)
	return c
}

// Get returns the instance of the service.
func Get(name string) interface{} {
	var (
		p  *service
		ok bool
	)
	EnterActionHook("get", name)
	for {
		serviceProviders.RLock()
		p, ok = serviceProviders.m[name]
		serviceProviders.RUnlock()
		if ok {
			p.RLock()
			if p.IsReady {
				s := p.Instance
				p.RUnlock()
				ExitActionHook("get", name)
				return s
			}
			p.RUnlock()
		}
		time.Sleep(10 * time.Millisecond) // XXX alternative with channels
	}
}

// List returns a list of currently ready to use services.
func List() map[string][]string {
	var (
		list = make(map[string][]string)
	)
	serviceProviders.RLock()
	for name, service := range serviceProviders.m {
		if service.IsReady {
			deps := []string{}
			for name := range service.Dependendents {
				deps = append(deps, name)
			}
			list[name] = deps
		}
	}
	serviceProviders.RUnlock()
	return list
}

// Fail signals the service that it is failed with anything
// critical. The service should listen on service.Failed() for these
// signals and handle them with the service restart. After restart
// service should Put() its instance to service again.
func Fail(name string) {
	serviceProviders.RLock()
	c, ok := serviceProviders.m[name]
	serviceProviders.RUnlock()
	if ok {
		EnterActionHook("failed", name)
		c.Lock()
		if !c.IsReady {
			c.Unlock()
			return
		}
		c.IsReady = false
		c.Unlock()
		c.RLock()
		for key := range c.Dependendents {
			if name == key {
				continue // cyclic dependency detected
			}
			go func(name string) {
				Fail(name)
			}(key)
		}
		c.NeedRestart <- struct{}{}
		c.RUnlock()
		ExitActionHook("failed", name)
	}
}

type service struct {
	Name        string
	NeedRestart chan struct{}

	sync.RWMutex
	IsReady       bool
	Instance      interface{}
	Dependendents map[string]bool // XXX сюда лучше каналы на рестарт сервисов
}

// WaitFor waits for a service and returns it instance.
func (c *service) WaitFor(name string) interface{} {
	var (
		p       *service
		ok, set bool
	)
	EnterActionHook("wait", name)
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
				s := p.Instance
				p.RUnlock()
				ExitActionHook("wait", name)
				return s
			}
			p.RUnlock()
		}
		time.Sleep(10 * time.Millisecond) // XXX alternative with channels
	}
}

// Ready puts the service instance into service structure and says the
// service is ready to serve.
func (c *service) Ready(service interface{}) {
	EnterActionHook("ready", c.Name)
	c.Lock()
	c.Instance = service
	c.IsReady = true
	c.Unlock()
	ExitActionHook("ready", c.Name)
}

// Failed notify the service that it is failed and should be
// restarted. The service should listen on service.Failed() for these
// signals and handle them with the service restart. After restart
// service should Put() its instance to service again.
func (c *service) Failed() chan struct{} { // XXX rename
	return c.NeedRestart
}
