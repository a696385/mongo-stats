package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"time"
)

type HostStatus struct {
	Host           Host
	Up             bool
	ConnectionTime time.Duration
}

func (this *HostStatus) Check() error {
	//Connect to MongoDB
	fmt.Printf("Connecting to %s:%d...  ", this.Host.Hostname, this.Host.Port)
	start := time.Now()
	session, err := mgo.Dial(this.Host.GetUrl())
	if err != nil {
		this.Up = false
		fmt.Println("ERROR - " + this.Host.GetUrl())
	} else {
		fmt.Println("OK")
		defer session.Close()
		end := time.Now()
		this.ConnectionTime = end.Sub(start)
		this.Up = true
	}
	return err
}
