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
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"gopkg.in/pipe.v2"
)

const (
	DB_USER     = "postgres"
	DB_PASSWORD = "12344321"
	DB_NAME     = "tiramisu"
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
		return false, fmt.Errorf("%v: cannot start cmd", err.Error())
	}

	// MUST BE goroutine!!
	go timedSIGTERM(cmd.Process, d)

	iopsJSONDecoder(rc)

	err = cmd.Wait()
	if err != nil {
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

func iopsJSONDecoder(rc io.ReadCloser) {
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
				fmt.Printf("arguments: count: %v\n%v %v\n", len(argsList), argsList[2], argsList[28])
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
}

func main() {
	dbinfo := fmt.Sprintf("user=postgres password=12344321 dbname=tiramisu sslmode=disable")
	db, err := sql.Open("postgres", dbinfo)
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

	RestartProcess(iopscmd, 8*time.Second)
}
