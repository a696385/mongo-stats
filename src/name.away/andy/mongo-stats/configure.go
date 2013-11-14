package main

import (
	"flag"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
)

type DBShard struct {
	Id   string `bson:"_id"`
	Host string
}

type Host struct {
	Hostname string
	Port     int
}

type Shard struct {
	Name  string
	Hosts []Host
}

func main() {
	//Parse arguments
	mongoUrl := flag.String("url", "mongodb://localhost/admin", "mongodb connection url")
	flag.Parse()

	session, err := mgo.Dial(*mongoUrl)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	shardsCollection := session.DB("config").C("shards")

	var dbShards []DBShard
	err = shardsCollection.Find(bson.M{}).Sort("-_id").All(&dbShards)
	if err != nil {
		panic(err)
	}

	shards := make([]Shard, len(dbShards))
	for i, item := range dbShards {
		hostsNames := strings.Split(strings.Split(item.Host, "/")[1], ",")
		hosts := make([]Host, len(hostsNames))
		for index, hostName := range hostsNames {
			parts := strings.Split(hostName, ":")
			port := 27017
			if len(parts) == 2 {
				port, err = strconv.Atoi(parts[1])
				if err != nil {
					port = 27017
				}
			}
			hosts[index] = Host{parts[0], port}
		}
		shards[i] = Shard{Name: item.Id, Hosts: hosts}
	}
	fmt.Printf("Found shards: %+v\n", shards)
}
