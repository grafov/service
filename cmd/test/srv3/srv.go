package srv3

import (
	"github.com/grafov/service"
	"github.com/grafov/service/cmd/test/srv1"
	"github.com/grafov/service/cmd/test/srv2"
)

type C struct {
	c1 *srv1.C
	c2 *srv2.C
}

func Run() {
	var c = new(C)
	p := service.Provide("srv3")
	for {
		c = &C{}
		c.c1 = p.WaitFor("srv1").(*srv1.C)
		c.c2 = p.WaitFor("srv2").(*srv2.C)
		p.Ready(c)
		println("srv3 has started")
		<-p.Failed()
		println("srv3 has failed")
	}
}
