package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func timedSIGTERM(p *os.Process, d time.Duration) {
	log.Println("couting down:", d)
	_ = <-time.After(d)
	log.Println("count finished, sending signal")
	err := p.Signal(syscall.SIGTERM)
	log.Println("signal sent")
	if err != nil {
		log.Panic(err)
	}
}

func SubRestartProcess(cmd *exec.Cmd, d time.Duration) (bool, error) {
	err := cmd.Start()
	if err != nil {
		//log.Fatalf("Error: [%v]\n", err)
		return false, fmt.Errorf("%v: cannot start cmd", err.Error())
	}

	// goroutine
	timedSIGTERM(cmd.Process, d)

	err = cmd.Wait()
	if err != nil {
		//log.Fatalf("Error: [%v]\n", err)
		return false, err
	}
	return cmd.ProcessState.Success(), nil
}

func RenewCmd(cmd *exec.Cmd) *exec.Cmd {
	var newCmd *exec.Cmd
	*newCmd = *cmd
	return newCmd
}

func RestartProcess(cmd *exec.Cmd, d time.Duration) {
	status := true
	var err error
	for status != false && err == nil {
		status, err = SubRestartProcess(cmd, d)
		log.Println("restarting...")
		log.Printf("status = %v, error = %v\n", status, err)
		if status == true {
			cmd = exec.Command(cmd.Path, cmd.Args[1])
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
	}
}

func main() {
	fmt.Print()
	iopscmd := exec.Command("stap", "iostat-json.stp")
	iopscmd.Stdout = os.Stdout
	iopscmd.Stderr = os.Stderr

	RestartProcess(iopscmd, 10*time.Second)

	// err := iopscmd.Start()
	// log.Println("pid:", iopscmd.Process.Pid)

	// timedSIGTERM(iopscmd.Process, 10*time.Second)

	// if err != nil {
	// 	panic(err)
	// }

	// err = iopscmd.Wait()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("process status : %v\n", iopscmd.ProcessState.Success())
}
