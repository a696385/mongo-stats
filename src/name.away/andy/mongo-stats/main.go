package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type Shard struct {
	Id   string `bson:"_id"`
	Host string
}

func main() {
	session, err := mgo.Dial()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	shardsCollection := session.DB("config").C("shards")

	var shards []Shard
	err = shardsCollection.Find(bson.M{}).Sort("-_id").All(&shards)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Shards: %+v\n", shards)
}
