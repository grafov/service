package srv2

import (
	"sync"

	"github.com/grafov/service"
	"github.com/grafov/service/cmd/test/srv1"
)

type C struct {
	sync.RWMutex
	c1 *srv1.C
}

func Run() {
	var c = new(C)
	p := service.Provide("srv2")
	for {
		c = &C{c1: p.WaitFor("srv1").(*srv1.C)}
		p.Ready(c)
		println("srv2 has started")
		<-p.Failed()
		println("srv2 has failed")
	}
}
