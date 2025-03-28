package main

import (
	"fmt"
	"os"

	"github.com/krbreyn/reminder/daemon"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("put usage info here")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "daemon":
		daemon.RunDaemon()

	case "add":

	case "delete":

	case "list":

	default:
		fmt.Printf("input not understood: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func addCmd() {

}

func deleteCmd() {

}

func listCmd() {
	file := daemon.OpenDataFile()
	defer file.Close()

}
