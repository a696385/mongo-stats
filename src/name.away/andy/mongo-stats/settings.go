package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Host struct {
	Hostname string
	Port     int
	Username string
	Password string
	Type     string
}

type Shard struct {
	Name  string
	Hosts Hosts
}

type Settings struct {
	Mongos Hosts
	Shards []Shard
}

func (this *Settings) Load(fileName string) error {
	file, e := ioutil.ReadFile(fileName)
	if e != nil {
		return e
	}
	e = json.Unmarshal(file, this)
	if e != nil {
		return e
	}
	return nil
}

func (this *Settings) Save(fileName string) error {
	data, e := json.Marshal(this)
	if e != nil {
		return e
	}
	e = ioutil.WriteFile(fileName, data, os.ModePerm)
	if e != nil {
		return e
	}
	return nil
}

type Hosts []Host

func (this *Host) GetUrl() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/admin", this.Username, this.Password, this.Hostname, this.Port)
}

func (this *Host) GetName() string {
	return fmt.Sprintf("%s:%d", this.Hostname, this.Port)
}

func (this Hosts) Less(i, j int) bool {
	return this[i].GetName() < this[j].GetName()
}

func (this Hosts) Len() int {
	return len(this)
}

func (this Hosts) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}
