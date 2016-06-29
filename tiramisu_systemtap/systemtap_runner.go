package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/Choestelus/furry-robot"
)

const MainPeriod = 10 * time.Second
const AssignPeriod = 10 * time.Second
const PostmarkPeriod = 20 * time.Minute

func RunTheCube(ticker *time.Ticker) {
	cubeCmd := exec.Command("python", "../tiramisu_src/line_model.py")
	var vms []corgis.TiramisuState

	for _ = range ticker.C {
		corgis.DB.Table("tiramisu_state").Select("vm_name").Find(&vms)
		// fmt.Printf("vms length: %v\n", len(vms))
		for _, e := range vms {
			newcmd := exec.Command(cubeCmd.Path, cubeCmd.Args[1], e.Name)
			//newcmd.Stdout = os.Stdout
			err := newcmd.Run()
			if err != nil {
				log.Printf("%v: linemodel error: %v\n", e.Name, err)
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

	go corgis.HttpServe()

	assignticker := time.NewTicker(AssignPeriod)
	cubeticker := time.NewTicker(MainPeriod)
	postmarkticker := time.NewTicker(PostmarkPeriod)

	go RunTheCube(cubeticker)

	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			corgis.AssignState()
		}
	}(assignticker)

	go func(ticker *time.Ticker) {
		for _ = range ticker.C {
			go corgis.CallPostmark("HDD")
			go corgis.CallPostmark("SSD")
		}
	}(postmarkticker)

	c := make(chan bool)
	var _ = <-c
	fmt.Println("-------------------------------------------------------------")
}
