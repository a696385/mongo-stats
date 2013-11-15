package main

import (
	"github.com/gonuts/commander"
	"github.com/gonuts/flag"
	"github.com/nsf/termbox-go"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"sort"
	"time"
)

type HostStatistic struct {
	Host      Host
	Insert    int
	Update    int
	Delete    int
	Query     int
	Command   int
	IsUp      bool
	LastError error
	IsMaster  bool
}

type DBOpcounters struct {
	Insert  int
	Update  int
	Delete  int
	Query   int
	Command int
}

type DBRepl struct {
	Ismaster bool
	Me       string
}

type DBServerStatus struct {
	Opcounters DBOpcounters
	Repl       DBRepl
}

var log string

func startMonitor(host *Host, report chan HostStatistic) {
	stat := HostStatistic{Host: *host}
	stat.IsUp = false
	tickChannel := time.Tick(time.Duration(1) * time.Second)
	session, err := mgo.Dial(host.GetUrl() + "?connect=direct")
	if err != nil {
		for {
			<-tickChannel
			session, err = mgo.Dial(host.GetUrl() + "?connect=direct")
			if err == nil {
				break
			} else {
				stat.LastError = err
				log = host.GetUrl() + " ERROR " + err.Error()
			}
		}
	}
	DB := session.DB("admin")

	lastOperations := DBOpcounters{-1, -1, -1, -1, -1}

	stat.IsUp = true
	for {
		<-tickChannel
		var result DBServerStatus
		err = DB.Run(bson.M{"serverStatus": 1}, &result)
		if err != nil {
			stat.LastError = err
			stat.IsUp = false
			continue
		}
		stat.LastError = nil
		stat.IsUp = false
		if lastOperations.Command >= 0 {
			stat.Command = result.Opcounters.Command - lastOperations.Command
			stat.Insert = result.Opcounters.Insert - lastOperations.Insert
			stat.Update = result.Opcounters.Update - lastOperations.Update
			stat.Delete = result.Opcounters.Delete - lastOperations.Delete
			stat.Query = result.Opcounters.Query - lastOperations.Query
		}
		lastOperations = result.Opcounters
		stat.IsMaster = result.Repl.Ismaster
		log = host.GetUrl() + " == " + result.Repl.Me
		report <- stat
	}

}

func termboxEvent(report chan termbox.Event, allowReport chan bool, quit chan bool) {
	for {
		select {
		case <-allowReport:
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				report <- ev
			}
		case <-quit:
			return
		}

	}
}

func executeMonitorCommand(cmd *commander.Command, args []string) {

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	err = termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	if err != nil {
		termbox.Close()
		panic(err)
	}
	w, h := termbox.Size()
	_ = h //debug
	w = (int)(float32(w) * 0.75)

	tabel := CreateTable()

	tabel.AddColumn("Name", 25)
	tabel.AddColumn("Command", 15)
	tabel.AddColumn("Query", 15)
	tabel.AddColumn("Insert", 15)
	tabel.AddColumn("Update", 15)
	tabel.AddColumn("Delete", 15)

	tabel.Print(0, 0, w, h)
	termbox.Flush()

	quit := make(chan bool)
	allowReport := make(chan bool)
	terminalEvents := make(chan termbox.Event)
	go termboxEvent(terminalEvents, allowReport, quit)

	settings := Settings{}
	settings.Load("settings.json")

	reportsChannel := make(chan HostStatistic)

	for _, shard := range settings.Shards {
		for i := 0; i < len(shard.Hosts); i++ {
			func(index int) {
				go startMonitor(&shard.Hosts[index], reportsChannel)
			}(i)
		}
	}

	reports := make(map[Host]HostStatistic)
	tickChannel := time.Tick(time.Duration(1) * time.Second)
	clearHostsListChannel := time.Tick(time.Duration(30) * time.Second)

	allowReport <- true
	for {
		select {
		case <-clearHostsListChannel:
			reports = make(map[Host]HostStatistic)
		case <-tickChannel:
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			tabel.RemoveRows()
			hosts := make(Hosts, len(reports))
			i := 0
			for key, _ := range reports {
				hosts[i] = key
				i++
			}
			sort.Sort(hosts)

			totalOperations := DBOpcounters{}
			for _, host := range hosts {
				prefix := ""
				report := reports[host]
				if report.IsMaster {
					prefix = "M:"
				}
				totalOperations.Command += report.Command
				totalOperations.Query += report.Query
				totalOperations.Insert += report.Insert
				totalOperations.Update += report.Update
				totalOperations.Delete += report.Delete
				tabel.AddRow(prefix+report.Host.GetName(), report.Command, report.Query, report.Insert, report.Update, report.Delete)
			}
			tabel.SetFooter("TOTAL", totalOperations.Command, totalOperations.Query, totalOperations.Insert, totalOperations.Update, totalOperations.Delete)
			tabel.Print(0, 0, w, h)

			x := 0
			for _, r := range log {
				termbox.SetCell(x, h-1, r, termbox.ColorDefault, termbox.ColorDefault)
				x++
			}
			termbox.Flush()

		case report := <-reportsChannel:
			reports[report.Host] = report

		case ev := <-terminalEvents:
			switch ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyEsc:
					quit <- true
					return
				default:
					allowReport <- true
				}
			default:
				allowReport <- true
			}

		}
	}
}

func getMonitorCommend() *commander.Command {
	cmd := &commander.Command{
		Run:       executeMonitorCommand,
		UsageLine: "monitor [options]",
		Short:     "monitoring of severs",
		Long:      "monitoring of severs",
		Flag:      *flag.NewFlagSet("monitor", flag.ExitOnError),
	}
	return cmd
}
