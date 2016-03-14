package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/Choestelus/furry-robot"
)

func main() {
	task := corgis.JobScheduler{}
	task.Cmd = exec.Command("stap", "iostat-json.stp")
	task.ExecPeriod = 8 * time.Second
	task.InitCmd()
	task.Type = corgis.Timed
	task.Muffled = true
	go task.Execute()

	task2 := corgis.JobScheduler{}
	task2.Cmd = exec.Command("stap", "latency_diskread.stp")
	task2.ExecPeriod = 8 * time.Second
	task2.InitCmd()
	task2.Type = corgis.Streaming
	task2.LType = corgis.LRead
	task2.Muffled = true
	go task2.Execute()

	task3 := corgis.JobScheduler{}
	task3.Cmd = exec.Command("stap", "latency_diskwrite.stp")
	task3.ExecPeriod = 8 * time.Second
	task3.InitCmd()
	task3.Type = corgis.Streaming
	task3.LType = corgis.LWrite
	task3.Muffled = true
	go task3.Execute()

	c := make(chan bool)
	var _ = <-c
	fmt.Println("-------------------------------------------------------------")
}
