# Dependency Tree for Services [![CircleCI](https://circleci.com/gh/grafov/service/tree/master.svg?style=svg)](https://circleci.com/gh/grafov/service/tree/master) [![GoDoc](https://godoc.org/github.com/grafov/service?status.svg)](https://godoc.org/github.com/grafov/service) 

_The package helps you separate constant running parts of code as
"services" where one "service" could be dependent of others._

You could run parts of logic inside your application as independent
"services", inside goroutines for example. "Services" in this context
are parts of the code inside application that could be dependable each
of another. The package helps track their dependencies and restart
them by chain.

## Features

The package does:

 * Provides registry for declaring parts of code as "services"
 * Allow setup which other services depends on a declared service
 * Notifies other services when the service they depended failed for some reason

## How it works

It is simplier explain in an example. Database client reads
configuration from etcd (or another kind of remote storage) and
initializes the client instance.  When remote configuration changed
(for example another database node added) you should reread it and
reinitialize client. With `service` you could run it like this (it is
pseudocode where only `service` calls are real):

```go

// read config from remote storage
go func() {
  for {
    s := service.Provide("configurator") // notifies the service registry that it provides "configurator"
    c := remoteConfig.ReadConfiguration()
    s.Ready(c) // indicate when service is ready and pushes `remoteConfig` object into service registry
    <-service.Failed() // waiting when the service failed or require reinitialization
  }
} 

// run database connector
go func() {
   for {
     s := service.Provide("dbclient") // nofifies the service registry that it provides "dbclient"
	 config := s.WaitFor("configurator") // it will wait until "configurator" is ready and gets it
     client := database.Connect(config.DBNodes)
	 <-service.Failed()
   }	 
} 
 
// Somewhere in the code:
service.Fail("configurator") // it will "fail" the "configurator" service and requires it's reinitialization

// It will fail also the depended services, "dbclient" in this
// example. So "dbclient" also get "fail" signal in the place where
// service.Failed() function waits and will pass to the next
// initialization loop.

// Anywhere in the code you can get current state of the object that provided by service.
// For example: 
dbclient := service.Get("dbclient").(database.ClientType) // Get() returns interface{} so type casting required
```

See more docs [![GoDoc](https://godoc.org/github.com/grafov/service?status.svg)](https://godoc.org/github.com/grafov/service)
