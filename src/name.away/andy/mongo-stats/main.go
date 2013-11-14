package main

import (
	"code.google.com/p/gopass"
	"flag"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
	"time"
)

type DBShard struct {
	Id   string `bson:"_id"`
	Host string
}

type DBMongoS struct {
	Id string `bson:"_id"`
}

var (
	username string
	password string
)

func main() {
	//Parse arguments
	mongoUrl := flag.String("url", "mongodb://localhost/admin", "mongodb connection url")
	isConfigure := flag.Bool("configure", false, "configure settigs")
	isCheckStatus := flag.Bool("check-status", false, "check servers status")
	flag.Parse()

	if *isConfigure {
		err := configure(*mongoUrl)
		if err != nil {
			panic(err)
		}
	} else if *isCheckStatus {
		err := checkStatus()
		if err != nil {
			panic(err)
		}
	} else {
		fmt.Println("Dont know to doing.... sorry")
	}
}

func checkStatus() error {
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
		statuses[index].Check()
		index++
	}
	for _, shard := range settings.Shards {
		for _, host := range shard.Hosts {
			statuses[index] = HostStatus{Host: host}
			statuses[index].Check()
			index++
		}
	}
	var (
		minConnectionTime     time.Duration
		minConnectionTimeHost Host
		maxConnectionTime     time.Duration
		maxConnectionTimeHost Host
	)
	for _, status := range statuses {
		if status.Up {
			if minConnectionTime == 0 || minConnectionTime > status.ConnectionTime {
				minConnectionTime = status.ConnectionTime
				minConnectionTimeHost = status.Host
			}
			if maxConnectionTime == 0 || maxConnectionTime < status.ConnectionTime {
				maxConnectionTime = status.ConnectionTime
				maxConnectionTimeHost = status.Host
			}
			fmt.Printf("%s:%d - Up %v\n", status.Host.Hostname, status.Host.Port, status.ConnectionTime)
		} else {
			fmt.Printf("%s:%d - DOWN\n", status.Host.Hostname, status.Host.Port)
		}
	}
	if minConnectionTime != 0 && maxConnectionTime != 0 {
		fmt.Printf("Max connection time on %s:%d - %v\n", maxConnectionTimeHost.Hostname, maxConnectionTimeHost.Port, maxConnectionTime)
		fmt.Printf("Min connection time on %s:%d - %v\n", minConnectionTimeHost.Hostname, minConnectionTimeHost.Port, minConnectionTime)
	}
	return nil
}

func configure(mongoUrl string) error {
	//Connect to MongoDB
	session, err := mgo.Dial(mongoUrl)
	if err != nil {
		return err
	}
	defer session.Close()

	//Find shards
	shardsCollection := session.DB("config").C("shards")
	var dbShards []DBShard
	err = shardsCollection.Find(bson.M{}).Sort("-_id").All(&dbShards)
	if err != nil {
		return err
	}

	shards := make([]Shard, len(dbShards))
	for i, item := range dbShards {
		hostsNames := strings.Split(strings.Split(item.Host, "/")[1], ",")
		hosts := getUsernamePasswords(hostsNames, "rs")
		shards[i] = Shard{Name: item.Id, Hosts: hosts}
	}

	//Find mongos servers
	mongosCollection := session.DB("config").C("mongos")
	var dbMongos []DBMongoS
	err = mongosCollection.Find(bson.M{}).Sort("-_id").All(&dbMongos)
	if err != nil {
		return err
	}

	hostsNames := make([]string, len(dbMongos))
	for i, item := range dbMongos {
		hostsNames[i] = item.Id
	}

	mongos := getUsernamePasswords(hostsNames, "mongos")

	settings := Settings{mongos, shards}
	err = settings.Save("settings.json")
	if err != nil {
		return err
	}
	return nil
}

func getUsernamePasswords(hosts []string, hostType string) []Host {
	var err error

	host := make([]Host, len(hosts))
	for i, s := range hosts {
		parts := strings.Split(s, ":")
		port := 27017
		if len(parts) == 2 {
			port, err = strconv.Atoi(parts[1])
			if err != nil {
				port = 27017
			}
		}
		fmt.Printf("\nAutorization for %s: \n", s)
		var newUserName string
		var newPassword string
		fmt.Printf("Username[default \"%s\"]: ", username)
		fmt.Scanf("%s", &newUserName)
		if newUserName != "" {
			username = newUserName
		}
		newPassword, err = gopass.GetPass("Password[Enter to tack previos]: ")
		if err != nil {
			panic(err)
		}
		if newPassword != "" {
			password = newPassword
		}
		host[i] = Host{parts[0], port, username, password, hostType}
	}
	return host
}
