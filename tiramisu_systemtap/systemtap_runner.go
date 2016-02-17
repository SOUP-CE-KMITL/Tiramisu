package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"gopkg.in/pipe.v2"
)

type ProcessLatency struct {
	Timestamp int64  `json:timestamp`
	Pid       int    `json:pid`
	PPid      int    `json:ppid`
	Execname  string `json:execname`
	Latency   int    `json:latency`
}

type ProcessIOPS struct {
	Pid        string `json:pid`
	Read       int    `json:read`
	ReadTotal  int    `json:read_total`
	ReadAvg    int    `json:read_avg`
	Write      int    `json:write`
	WriteTotal int    `json:write_total`
	WriteAvg   int    `json:write_avg`
}

type Pair struct {
	Value int
	Count int
}

func GetArguments(pid int) []string {
	if pid == 0 {
		return nil
	}
	filename := "/proc/" + strconv.Itoa(pid) + "/cmdline"
	p := pipe.Line(
		pipe.ReadFile(filename),
		pipe.Exec("strings", "-1"),
	)
	output, err := pipe.CombinedOutput(p)
	if err != nil {
		fmt.Printf("error:[%v]\n", err)
	}
	return strings.Fields(string(output))
}

func timedSIGTERM(p *os.Process, d time.Duration) {
	_ = <-time.After(d)
	err := p.Signal(syscall.SIGTERM)
	if err != nil {
		log.Panic(err)
	}
}

func ProbeLatency(cmd *exec.Cmd, d time.Duration, ch chan Pair) {
	latencyPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}

	latencyJSONDecoder(latencyPipe, d, ch)

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
}

func latencyJSONDecoder(rc io.ReadCloser, d time.Duration, c chan Pair) {
	latencyDecoder := json.NewDecoder(rc)
	openToken, err := latencyDecoder.Token()
	if err != nil {
		log.Fatalf("error: %v/n", err)
	}
	var _ = openToken
	cumulativeLatency := 0
	cumulativeLatencyCount := 0

	// dispatcher & resetter
	var mutex = &sync.Mutex{}
	ticker := time.NewTicker(d)
	go func(c chan Pair) {
		for _ = range ticker.C {
			mutex.Lock()
			// log.Printf("latency, count: [%v] [%v]\n", cumulativeLatency, cumulativeLatencyCount)
			c <- Pair{Value: cumulativeLatency, Count: cumulativeLatencyCount}
			cumulativeLatency = 0
			cumulativeLatencyCount = 0
			mutex.Unlock()
		}
	}(c)

	for latencyDecoder.More() {
		var message ProcessLatency
		err := latencyDecoder.Decode(&message)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		if message.Execname == "qemu-kvm" {
			//fmt.Printf("PID: [%v], name: [%v], latency: [%v]\n", message.Pid, message.Execname, message.Latency)

			cumulativeLatencyCount += 1
			cumulativeLatency += message.Latency
		}
	}
	closeToken, err := latencyDecoder.Token()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	var _ = closeToken
}

func SubRestartProcess(cmd *exec.Cmd, d time.Duration, rc io.ReadCloser, cHDD chan Pair, cSSD chan Pair) (bool, error) {
	err := cmd.Start()
	if err != nil {
		return false, fmt.Errorf("%v: cannot start cmd", err.Error())
	}

	// MUST BE goroutine!!
	go timedSIGTERM(cmd.Process, d)

	iopsJSONDecoder(rc, cHDD, cSSD)

	err = cmd.Wait()
	if err != nil {
		return false, err
	}
	return cmd.ProcessState.Success(), nil
}

func RestartProcess(cmd *exec.Cmd, d time.Duration, cHDD chan Pair, cSSD chan Pair) {
	status := true
	var err error
	var iopsPipe io.ReadCloser
	for status != false && err == nil {
		if status == true {
			cmd = exec.Command(cmd.Path, cmd.Args[1])
			cmd.Stderr = os.Stderr
			iopsPipe, err = cmd.StdoutPipe()
			if err != nil {
				log.Fatalf("error %v\n", err)
			}
		}
		status, err = SubRestartProcess(cmd, d, iopsPipe, cHDD, cSSD)
		// log.Println("restarting...")
		// log.Printf("status = %v, error = %v\n", status, err)
	}
}

func iopsJSONDecoder(rc io.ReadCloser, cHDD chan Pair, cSSD chan Pair) {
	iopsDecoder := json.NewDecoder(rc)
	openToken, err := iopsDecoder.Token()
	if err != nil {
		log.Fatalf("error reading openToken: %v\n", err)
	}
	// fmt.Printf("%v %T\n", openToken, openToken)
	var _ = openToken

	cumulativeIOPSSSD := 0
	cumulativeIOPSSSDCount := 0
	cumulativeIOPSHDD := 0
	cumulativeIOPSHDDCount := 0
	for iopsDecoder.More() {
		var message ProcessIOPS
		err := iopsDecoder.Decode(&message)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("PID: [%v], IO Read: [%v] IO Write [%v]\n", message.Pid, message.Read, message.Write)
		var PIDNumber int
		if message.Pid != "" {
			PIDNumber, err = strconv.Atoi(message.Pid)
		}
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}
		argsList := GetArguments(PIDNumber)
		if len(argsList) != 0 {
			if argsList[0] == "/usr/libexec/qemu-kvm" {
				if strings.Contains(argsList[28], "SSD") {
					cumulativeIOPSSSDCount += 1
					cumulativeIOPSSSD += message.Read + message.Write
				} else {
					cumulativeIOPSHDDCount += 1
					cumulativeIOPSHDD += message.Read + message.Write
				}
				// fmt.Printf("PID: [%v], IO Read: [%v] IO Write [%v]\n", message.Pid, message.Read, message.Write)
				// fmt.Printf("arguments: count: %v\n%v %v\n", len(argsList), argsList[2], argsList[28])
				// fmt.Printf("it is ssd?: %v\n", strings.Contains(argsList[28], "SSD"))
			}
		}
	}

	closeToken, err := iopsDecoder.Token()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%v %T\n", closeToken, closeToken)
	var _ = closeToken
	fmt.Printf("SSD: cIOPS = %v cIOPSCount = %v\n", cumulativeIOPSSSD, cumulativeIOPSSSDCount)
	fmt.Printf("HDD: cIOPS = %v cIOPSCount = %v\n", cumulativeIOPSHDD, cumulativeIOPSHDDCount)
	cSSD <- Pair{Value: cumulativeIOPSSSD, Count: cumulativeIOPSSSDCount}
	cHDD <- Pair{Value: cumulativeIOPSHDD, Count: cumulativeIOPSHDDCount}
}

func main() {
	dbinfo := fmt.Sprintf("user=postgres password=12344321 dbname=tiramisu sslmode=disable")
	db, err := sql.Open("postgres", dbinfo)

	IOPSSSDchan := make(chan Pair)
	IOPSHDDchan := make(chan Pair)
	latencyReadChan := make(chan Pair)
	latencyWriteChan := make(chan Pair)

	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error: Could not establish a connection with the database %v\n", err)
	}
	rows, err := db.Query(`SELECT * FROM tiramisu_state`)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var vm_name string
		var latency float64
		var iops float64
		var latency_hdd float64
		var iops_hdd float64
		var latency_ssd float64
		var iops_ssd float64

		err := rows.Scan(&vm_name, &latency, &iops, &latency_hdd, &iops_hdd, &latency_ssd, &iops_ssd)
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		fmt.Printf("%12v | %8v | %8v | %8v | %8v | %8v | %8v\n", vm_name, latency, iops, latency_hdd, iops_hdd, latency_ssd, iops_ssd)
	}

	fmt.Print()
	iopscmd := exec.Command("stap", "iostat-json.stp")
	latencyReadCmd := exec.Command("stap", "latency_diskread.stp")
	latencyWriteCmd := exec.Command("stap", "latency_diskwrite.stp")

	go RestartProcess(iopscmd, 8*time.Second, IOPSHDDchan, IOPSSSDchan)
	go ProbeLatency(latencyReadCmd, 8*time.Second, latencyReadChan)
	go ProbeLatency(latencyWriteCmd, 8*time.Second, latencyWriteChan)
	for {
		select {
		case x := <-IOPSHDDchan:
			fmt.Printf("iops hdd: c = %v v = %v\n", x.Count, x.Value)
		case x := <-IOPSSSDchan:
			fmt.Printf("iops ssd: c = %v v = %v\n", x.Count, x.Value)
		case x := <-latencyReadChan:
			fmt.Printf("latency read c = %v v = %v\n", x.Count, x.Value)
		case x := <-latencyWriteChan:
			fmt.Printf("latency write c = %v v = %v\n", x.Count, x.Value)
		}
	}
	var _ = iopscmd
}
