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
	Hosts []Host
}

type Settings struct {
	Mongos []Host
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

func (this *Host) GetUrl() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/admin", this.Username, this.Password, this.Hostname, this.Port)
}
