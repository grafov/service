package srv1

import (
	"sync"

	"github.com/grafov/service"
)

type C struct {
	sync.RWMutex
}

func Run() {
	p := service.Provide("srv1")
	for {
		c := new(C)
		p.Ready(c)
		println("srv1 has started")
		<-p.Failed()
		println("srv1 has failed")
	}
}
