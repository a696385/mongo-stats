package main

import (
	"fmt"
	"github.com/aybabtme/color/brush"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"labix.org/v2/mgo"
	"strconv"
	"time"
)

type HostStatus struct {
	Host           Host
	Up             bool
	ConnectionTime time.Duration
}

func (this *HostStatus) Check() error {
	//Connect to MongoDB
	start := time.Now()
	session, err := mgo.Dial(this.Host.GetUrl())
	if err != nil {
		this.Up = false
	} else {
		defer session.Close()
		end := time.Now()
		this.ConnectionTime = end.Sub(start)
		this.Up = true
	}
	return err
}

func executeChechStatusCommand(cmd *commander.Command, args []string) {
	settings := Settings{}
	settings.Load("settings.json")

	serversCount := len(settings.Mongos)
	for _, item := range settings.Shards {
		serversCount += len(item.Hosts)
	}

	index := 0
	statuses := make([]HostStatus, serversCount)
	for _, host := range settings.Mongos {
		statuses[index] = HostStatus{Host: host}
		index++
	}
	for _, shard := range settings.Shards {
		for _, host := range shard.Hosts {
			statuses[index] = HostStatus{Host: host}
			index++
		}
	}
	c := make(chan error, serversCount)
	for i := 0; i < serversCount; i++ {
		go func(status *HostStatus, c chan error) {
			c <- status.Check()
		}(&statuses[i], c)
	}
	errors := 0
	completed := 0
	tickChannel := time.Tick(time.Duration(1) * time.Second)
	for {
		if completed >= serversCount {
			break
		}
		select {
		case <-tickChannel:
			fmt.Printf("Completed: %d of %d\n", completed, serversCount)
		case err := <-c:
			if err != nil {
				errors++
			}
			completed++
		}
	}
	fmt.Printf("Errors: %s\n", brush.DarkRed(strconv.Itoa(errors)))
	var (
		minConnectionTime time.Duration
		maxConnectionTime time.Duration
	)
	for _, status := range statuses {
		if status.Up {
			if minConnectionTime == 0 || minConnectionTime > status.ConnectionTime {
				minConnectionTime = status.ConnectionTime
			}
			if maxConnectionTime == 0 || maxConnectionTime < status.ConnectionTime {
				maxConnectionTime = status.ConnectionTime
			}
		}
	}
	for _, status := range statuses {
		var serverName = status.Host.GetName()
		if status.Up {
			connectionTimeColored := fmt.Sprintf("%v", status.ConnectionTime)
			if status.ConnectionTime == minConnectionTime {
				connectionTimeColored = fmt.Sprintf("%s", brush.Green(connectionTimeColored))
			} else if status.ConnectionTime == maxConnectionTime {
				connectionTimeColored = fmt.Sprintf("%s", brush.Red(connectionTimeColored))
			}
			fmt.Printf("%s - %s\n", brush.DarkGreen(serverName), connectionTimeColored)

		} else {
			fmt.Printf("%s - DOWN\n", brush.DarkRed(serverName))
		}
	}
}

func getCheckStatusCommend() *commander.Command {
	cmd := &commander.Command{
		Run:       executeChechStatusCommand,
		UsageLine: "check-status [options]",
		Short:     "check status of servers",
		Long:      "check status of servers",
		Flag:      *flag.NewFlagSet("check-status", flag.ExitOnError),
	}
	return cmd
}
