package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Choestelus/furry-robot"
)

const MainPeriod = 10 * time.Second

func RunTheCube(ticker *time.Ticker) {
	cubeCmd := exec.Command("python", "../tiramisu_src/the_cube.py")
	var vms []corgis.TiramisuState

	for _ = range ticker.C {
		corgis.DB.Table("tiramisu_state").Select("vm_name").Find(&vms)
		log.Println("ticked cube")
		fmt.Printf("vms length: %v\n", len(vms))
		for i, e := range vms {
			fmt.Printf("cube [%v]: %v\n:", i, e.Name)
			newcmd := exec.Command(cubeCmd.Path, cubeCmd.Args[1], e.Name)
			newcmd.Stdout = os.Stdout
			err := newcmd.Run()
			if err != nil {
				log.Printf("cube error: %v\n", err)
			}
		}
	}
	fmt.Printf("should never been here\n")
}

func main() {
	task := corgis.JobScheduler{}
	task.Cmd = exec.Command("stap", "iostat-json.stp")
	task.ExecPeriod = MainPeriod
	task.InitCmd()
	task.Type = corgis.Timed
	task.Muffled = true
	go task.Execute()

	task2 := corgis.JobScheduler{}
	task2.Cmd = exec.Command("stap", "latency_diskread.stp")
	task2.ExecPeriod = MainPeriod
	task2.InitCmd()
	task2.Type = corgis.Streaming
	task2.LType = corgis.LRead
	task2.Muffled = true
	go task2.Execute()

	task3 := corgis.JobScheduler{}
	task3.Cmd = exec.Command("stap", "latency_diskwrite.stp")
	task3.ExecPeriod = MainPeriod
	task3.InitCmd()
	task3.Type = corgis.Streaming
	task3.LType = corgis.LWrite
	task3.Muffled = true
	go task3.Execute()

	assignticker := time.NewTicker(MainPeriod)
	cubeticker := time.NewTicker(MainPeriod)

	go RunTheCube(cubeticker)

	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			corgis.AssignAverage()
		}
	}(assignticker)

	c := make(chan bool)
	var _ = <-c
	fmt.Println("-------------------------------------------------------------")
}
