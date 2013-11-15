package main

import (
	"code.google.com/p/gopass"
	"fmt"
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"strconv"
	"strings"
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

func executeConfigureCommand(cmd *commander.Command, args []string) {
	mongoUrl := cmd.Flag.Lookup("url").Value.Get().(string)
	err := configure(mongoUrl)
	if err != nil {
		panic(err)
	}
}

func getConfgigureCommend() *commander.Command {
	cmd := &commander.Command{
		Run:       executeConfigureCommand,
		UsageLine: "configure [options]",
		Short:     "get servers list",
		Long:      "get servers list and accept autorisation",
		Flag:      *flag.NewFlagSet("configure", flag.ExitOnError),
	}
	cmd.Flag.String("url", "mongodb://localhost", "MongoDB Connection URL")
	return cmd
}
