package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type ProcessIOPS struct {
	Pid        string `json:pid`
	Read       int    `json:read`
	ReadTotal  int    `json:read_total`
	ReadAvg    int    `json:read_avg`
	Write      int    `json:write`
	WriteTotal int    `json:write_total`
	WriteAvg   int    `json:write_avg`
}

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

func SubRestartProcess(cmd *exec.Cmd, d time.Duration, rc io.ReadCloser) (bool, error) {
	err := cmd.Start()
	if err != nil {
		//log.Fatalf("Error: [%v]\n", err)
		return false, fmt.Errorf("%v: cannot start cmd", err.Error())
	}

	// goroutine
	go timedSIGTERM(cmd.Process, d)

	foo(rc)

	err = cmd.Wait()
	if err != nil {
		//log.Fatalf("Error: [%v]\n", err)
		return false, err
	}
	return cmd.ProcessState.Success(), nil
}

func RestartProcess(cmd *exec.Cmd, d time.Duration) {
	status := true
	var err error
	var iopsPipe io.ReadCloser
	for status != false && err == nil {
		if status == true {
			cmd = exec.Command(cmd.Path, cmd.Args[1])
			// cmd.Stdout = foo
			cmd.Stderr = os.Stderr
			iopsPipe, err = cmd.StdoutPipe()
			if err != nil {
				log.Fatalf("error %v\n", err)
			}
		}
		status, err = SubRestartProcess(cmd, d, iopsPipe)
		log.Println("restarting...")
		log.Printf("status = %v, error = %v\n", status, err)
	}
}

func foo(rc io.ReadCloser) {
	//rc := strings.NewReader(op)
	iopsDecoder := json.NewDecoder(rc)
	openToken, err := iopsDecoder.Token()
	if err != nil {
		log.Fatalf("error reading openToken: %v\n", err)
	}
	fmt.Printf("%v %T\n", openToken, openToken)

	for iopsDecoder.More() {
		var message ProcessIOPS
		err := iopsDecoder.Decode(&message)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("PID: [%v], IO Read: [%v] IO Write [%v]\n", message.Pid, message.Read, message.Write)
	}

	closeToken, err := iopsDecoder.Token()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v %T\n", closeToken, closeToken)
}

func main() {
	fmt.Print()
	iopscmd := exec.Command("stap", "iostat-json.stp")

	RestartProcess(iopscmd, 4*time.Second)

	// iopsDecoder := json.NewDecoder(iopsPipe)
	// openToken, err := iopsDecoder.Token()
	// if err != nil {
	// 	log.Fatalf("error reading openToken: %v\n", err)
	// }
	// fmt.Printf("%v %T\n", openToken, openToken)

	// for iopsDecoder.More() {
	// 	var message ProcessIOPS
	// 	err := iopsDecoder.Decode(&message)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	fmt.Printf("PID: [%v], IO Read: [%v] IO Write [%v]\n", message.Pid, message.Read, message.Write)
	// }

	// closeToken, err := iopsDecoder.Token()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%v %T\n", closeToken, closeToken)
}
