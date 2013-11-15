package main

import (
	"fmt"
	"github.com/gonuts/commander"
	"os"
)

var g_cmd *commander.Commander

func init() {
	g_cmd = &commander.Commander{
		Name: os.Args[0],
		Commands: []*commander.Command{
			getCheckStatusCommend(),
			getConfgigureCommend(),
			getMonitorCommend(),
		},
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("Try help...")
		return
	}

	err := g_cmd.Run(os.Args[1:])
	if err != nil {
		fmt.Printf("**err**: %v\n", err)
		os.Exit(1)
	}

}
