package srv4

import (
	"github.com/grafov/service"
	"github.com/grafov/service/cmd/test/srv2"
	"github.com/grafov/service/cmd/test/srv3"
)

type C struct {
	c2 *srv2.C
	c3 *srv3.C
}

func Run() {
	var c = new(C)
	p := service.Provide("srv4")
	for {
		c = &C{}
		c.c2 = p.WaitFor("srv2").(*srv2.C)
		c.c3 = p.WaitFor("srv3").(*srv3.C)
		p.Ready(c)
		println("srv4 has started")
		<-p.Failed()
		println("srv4 has failed")
	}
}
