package main

import (
	"fmt"
	"time"

	"github.com/grafov/service"
	"github.com/grafov/service/cmd/test/srv1"
	"github.com/grafov/service/cmd/test/srv2"
	"github.com/grafov/service/cmd/test/srv3"
	"github.com/grafov/service/cmd/test/srv4"
)

func main() {
	service.EnterActionHook = func(action, name string) { fmt.Printf("ENTER %s: %s %s\n", time.Now(), action, name) }
	service.ExitActionHook = func(action, name string) { fmt.Printf("EXIT %s: %s %s\n", time.Now(), action, name) }
	fmt.Println(service.List())
	go srv4.Run()
	go srv2.Run()
	go srv1.Run()
	time.Sleep(50 * time.Millisecond)
	fmt.Println(service.List())
	go srv3.Run()
	time.Sleep(500 * time.Millisecond)
	fmt.Println(service.List())
	service.Fail("srv1")
	service.Fail("srv3")
	time.Sleep(500 * time.Millisecond)
	fmt.Println(service.List())
	time.Sleep(900 * time.Millisecond)
	fmt.Println(service.List())
}
