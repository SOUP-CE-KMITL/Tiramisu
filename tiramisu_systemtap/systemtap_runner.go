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

var mutex = &sync.RWMutex{}
var wg sync.WaitGroup

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
		// fmt.Printf("error:[%v]\n", err)
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

func ProbeLatency(cmd *exec.Cmd, d time.Duration, ch chan Pair, cVM chan VMInformation) {
	//defer wg.Done()

	newcmd := exec.Command(cmd.Path, cmd.Args[1])
	latencyPipe, err := newcmd.StdoutPipe()
	if err != nil {
		log.Fatalf("probe latency error: %v\n", err)
	}
	err = newcmd.Start()
	if err != nil {
		log.Fatalf("probestart error: %v\n", err)
	}

	go timedSIGTERM(newcmd.Process, d*time.Second)
	latencyJSONDecoder(latencyPipe, d, ch, cVM)

	err = newcmd.Wait()
	if err != nil {
		log.Fatalf("cmdWait error: %v\n", err)
	}
}

func latencyJSONDecoder(rc io.ReadCloser, d time.Duration, c chan Pair, cVM chan VMInformation) {
	latencyDecoder := json.NewDecoder(rc)
	openToken, err := latencyDecoder.Token()
	if err != nil {
		log.Fatalf("openToken error: %v/n", err)
	}
	var _ = openToken
	cumulativeLatency := 0
	cumulativeLatencyCount := 0

	var vmList []VMInformation

	// dispatcher & resetter
	ticker := time.NewTicker(d)
	go func(c chan Pair) {
		for _ = range ticker.C {
			// log.Printf("latency, count: [%v] [%v]\n", cumulativeLatency, cumulativeLatencyCount)
			mutex.RLock()
			c <- Pair{Value: cumulativeLatency, Count: cumulativeLatencyCount}
			mutex.RUnlock()
			mutex.RLock()
			for _, elem := range vmList {
				cVM <- elem
			}
			mutex.RUnlock()
			mutex.Lock()
			vmList = vmList[:0]
			cumulativeLatency = 0
			cumulativeLatencyCount = 0
			mutex.Unlock()
		}
	}(c)

	for latencyDecoder.More() {
		var message ProcessLatency
		err := latencyDecoder.Decode(&message)
		if err != nil {
			log.Printf("latency decode error: %v\n", err)
		}
		if message.Execname == "qemu-kvm" {
			argsList := GetArguments(message.Pid)
			is_ssd := strings.Contains(argsList[28], "SSD")
			// fmt.Printf("PID: [%v], name: [%v], latency: [%v] ssd?[%v]\n", message.Pid, argsList[2], message.Latency, is_ssd)

			mutex.Lock()
			cumulativeLatencyCount += 1
			cumulativeLatency += message.Latency
			mutex.Unlock()

			// HERE HERE
			mutex.Lock()
			vmList = append(vmList, VMInformation{
				Name:    argsList[2],
				Latency: message.Latency,
				ISSSD:   is_ssd,
			})
			mutex.Unlock()
		}
	}
	closeToken, err := latencyDecoder.Token()
	if err != nil {
		log.Printf("closeToken error: %v\n", err)
	}
	var _ = closeToken
}

func SubRestartProcess(cmd *exec.Cmd, d time.Duration, rc io.ReadCloser, cHDD chan Pair, cSSD chan Pair, cVM chan VMInformation) (bool, error) {
	err := cmd.Start()
	if err != nil {
		return false, fmt.Errorf("%v: cannot start cmd", err.Error())
	}

	// MUST BE goroutine!!
	go timedSIGTERM(cmd.Process, d)

	iopsJSONDecoder(rc, cHDD, cSSD, cVM)

	err = cmd.Wait()
	if err != nil {
		return false, err
	}
	return cmd.ProcessState.Success(), nil
}

func RestartProcess(cmd *exec.Cmd, d time.Duration, cHDD chan Pair, cSSD chan Pair, cVM chan VMInformation) {
	status := true
	var err error
	var iopsPipe io.ReadCloser
	for status != false && err == nil {
		if status == true {
			cmd = exec.Command(cmd.Path, cmd.Args[1])
			cmd.Stderr = os.Stderr
			iopsPipe, err = cmd.StdoutPipe()
			if err != nil {
				log.Fatalf("thiserror %v\n", err)
			}
		}
		status, err = SubRestartProcess(cmd, d, iopsPipe, cHDD, cSSD, cVM)
		//log.Println("restarting...")
		//log.Printf("status = %v, error = %v\n", status, err)
	}
}

func RunProcessIOPS(cmd *exec.Cmd, d time.Duration, cHDD chan Pair, cSSD chan Pair, cVM chan VMInformation) {
	//defer wg.Done()
	var iopsPipe io.ReadCloser
	newcmd := exec.Command(cmd.Path, cmd.Args[1])
	newcmd.Stderr = os.Stderr
	iopsPipe, err := newcmd.StdoutPipe()
	if err != nil {
		log.Fatalf("the this error %v\n", err)
	}
	_, err = SubRestartProcess(newcmd, d, iopsPipe, cHDD, cSSD, cVM)
	if err != nil {
		log.Fatalf("thiserr %v\n", err)
	}
}

func iopsJSONDecoder(rc io.ReadCloser, cHDD chan Pair, cSSD chan Pair, cVM chan VMInformation) {
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

	var vmList []VMInformation

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
			log.Fatalf("iopsdecoder Error: %v\n", err)
		}
		argsList := GetArguments(PIDNumber)
		if len(argsList) != 0 {
			// fmt.Printf("%v\n", argsList)
			if argsList[0] == "/usr/libexec/qemu-kvm" {
				if strings.Contains(argsList[28], "SSD") {
					cumulativeIOPSSSDCount += 1
					cumulativeIOPSSSD += message.Read + message.Write
				} else {
					cumulativeIOPSHDDCount += 1
					cumulativeIOPSHDD += message.Read + message.Write
				}
				is_ssd := strings.Contains(argsList[28], "SSD")
				// fmt.Printf("PID: [%v], IO Read: [%v] IO Write [%v]\n", message.Pid, message.Read, message.Write)
				// fmt.Printf("arguments: count: %v\n%v %v\n", len(argsList), argsList[2], argsList[28])
				// fmt.Printf("it is ssd?: %v\n", is_ssd)

				// Here Here
				vmList = append(vmList,
					VMInformation{
						Name:  argsList[2],
						IOPS:  (message.Read + message.Write),
						ISSSD: is_ssd,
					})

			}
		}
	}

	closeToken, err := iopsDecoder.Token()
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("%v %T\n", closeToken, closeToken)
	var _ = closeToken
	// fmt.Printf("SSD: cIOPS = %v cIOPSCount = %v\n", cumulativeIOPSSSD, cumulativeIOPSSSDCount)
	// fmt.Printf("HDD: cIOPS = %v cIOPSCount = %v\n", cumulativeIOPSHDD, cumulativeIOPSHDDCount)
	cSSD <- Pair{Value: cumulativeIOPSSSD, Count: cumulativeIOPSSSDCount}
	cHDD <- Pair{Value: cumulativeIOPSHDD, Count: cumulativeIOPSHDDCount}
	for _, elem := range vmList {
		cVM <- elem
	}
}

func main() {
	vmInfos := make(map[string]VMInformation)
	dbinfo := fmt.Sprintf("user=postgres password=12344321 dbname=tiramisu sslmode=disable")
	db, err := sql.Open("postgres", dbinfo)

	IOPSSSDchan := make(chan Pair, 1)
	IOPSHDDchan := make(chan Pair, 1)
	latencyReadChan := make(chan Pair, 1)
	latencyWriteChan := make(chan Pair, 1)
	vmIOPSInfoChan := make(chan VMInformation, 1)
	vmLatencyReadInfoChan := make(chan VMInformation, 1)
	vmLatencyWriteInfoChan := make(chan VMInformation, 1)

	wait := make(chan bool)

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

	fmt.Printf("%12v | %12v | %12v | %12v | %12v | %12v | %12v\n", "vm_name", "latency", "iops", "latency_hdd", "iops_hdd", "latency_ssd", "iops_ssd")
	for rows.Next() {
		var vm_name string
		var latency float64
		var iops float64
		var latency_hdd float64
		var iops_hdd float64
		var latency_ssd float64
		var iops_ssd float64
		var t1 *time.Time
		var t2 *time.Time
		var t3 *time.Time

		err := rows.Scan(&vm_name, &latency, &iops, &latency_hdd, &iops_hdd, &latency_ssd, &iops_ssd, &t1, &t2, &t3)
		if err != nil {
			log.Fatalf("readrow error: %v\n", err)
		}
		fmt.Printf("%12v | %12v | %12v | %12v | %12v | %12v | %12v\n", vm_name, latency, iops, latency_hdd, iops_hdd, latency_ssd, iops_ssd)
	}

	fmt.Print()
	iopscmd := exec.Command("stap", "iostat-json.stp")
	latencyReadCmd := exec.Command("stap", "latency_diskread.stp")
	latencyWriteCmd := exec.Command("stap", "latency_diskwrite.stp")
	thecubeCmd := exec.Command("python", "../tiramisu_src/the_cube.py")

	var _ = iopscmd
	var _ = latencyReadCmd
	var _ = latencyWriteCmd
	var _ = thecubeCmd
	//wg.Add(1)
	go RunProcessIOPS(iopscmd, 8*time.Second, IOPSHDDchan, IOPSSSDchan, vmIOPSInfoChan)
	//wg.Add(1)
	go ProbeLatency(latencyReadCmd, 8*time.Second, latencyReadChan, vmLatencyReadInfoChan)
	//wg.Add(1)
	go ProbeLatency(latencyWriteCmd, 8*time.Second, latencyWriteChan, vmLatencyWriteInfoChan)
	//wg.Wait()
	// cubeticker := time.NewTicker(10 * time.Second)
	// go func() {
	// 	for _ = range cubeticker.C {
	// 		for _, e := range vmInfos {
	// 			fmt.Printf("--->%v\n", e.Name)
	// 			fmt.Printf("--->%v\n", thecubeCmd.Args)
	// 			fmt.Printf("--->%v\n", thecubeCmd.Path)
	// 			newcmd := exec.Command(thecubeCmd.Path, thecubeCmd.Args[1], e.Name)
	// 			newcmd.Stdout = os.Stdout
	// 			err := newcmd.Run()
	// 			if err != nil {
	// 				log.Printf("cube error %v\n", err)
	// 			}
	// 		}
	go func() {
		for {
			select {
			case x := <-IOPSHDDchan:
				fmt.Printf("iops hdd: c = %v v = %v\n", x.Count, x.Value)

				tx, err := db.Begin()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}
				defer tx.Rollback()
				stmt, err := tx.Prepare(`update tiramisu_state set iops_hdd=$1 where vm_name=$2`)
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}
				defer stmt.Close()

				for _, e := range vmInfos {
					if e.ISSSD {
						tmp := vmInfos[e.Name]
						if x.Count != 0 {
							tmp.IOPSHDD = x.Value / x.Count
						}
						vmInfos[e.Name] = tmp

						_, err := stmt.Exec(float64(vmInfos[e.Name].IOPSHDD), e.Name)
						//fmt.Printf("[[%v]]\n", res)
						if err != nil {
							log.Fatalf("IOPSHDDchan db error: %v\n", err)
						}
					}
				}
				err = tx.Commit()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}

			case x := <-IOPSSSDchan:
				fmt.Printf("iops ssd: c = %v v = %v\n", x.Count, x.Value)

				tx, err := db.Begin()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}
				defer tx.Rollback()
				stmt, err := tx.Prepare(`update tiramisu_state set iops_ssd=$1 where vm_name=$2`)
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}
				defer stmt.Close()

				for _, e := range vmInfos {
					if !e.ISSSD {
						tmp := vmInfos[e.Name]
						if x.Count != 0 {
							tmp.IOPSSSD = x.Value / x.Count
						}
						vmInfos[e.Name] = tmp

						_, err := stmt.Exec(float64(vmInfos[e.Name].IOPSSSD), e.Name)
						//fmt.Printf("[[%v]]\n", res)
						if err != nil {
							log.Fatalf("IOPSSDchan db error: %v\n", err)
						}
					}
				}

				err = tx.Commit()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}

			case x := <-vmIOPSInfoChan:
				// If exist
				if _, ok := vmInfos[x.Name]; ok {
					tmp := vmInfos[x.Name]
					tmp.IOPS = x.IOPS
					tmp.ISSSD = x.ISSSD

					if vmInfos[x.Name].ISSSD {
						tmp.IOPSSSD = x.IOPSSSD
					} else {
						tmp.IOPSHDD = x.IOPS
					}
					vmInfos[x.Name] = tmp

				} else {
					vmInfos[x.Name] = VMInformation{
						Name:  x.Name,
						IOPS:  x.IOPS,
						ISSSD: x.ISSSD,
					}
				}
				fmt.Printf("--> [%v]\n", vmInfos[x.Name])
				txn, err := db.Begin()
				if err != nil {
					log.Fatalf("dberr: %v\n", err)
				}
				defer txn.Rollback()
				stmt, err := txn.Prepare(`update tiramisu_state set iops=$1 where vm_name=$2`)
				if err != nil {
					log.Fatalf("dberror: %v\n", err)
				}
				// defer stmt.Close()
				_, err = stmt.Exec(float64(vmInfos[x.Name].IOPS), x.Name)
				// fmt.Printf("[[%v]]\n", res)
				if err != nil {
					panic(err)
				}
				// err = stmt.Close()
				// if err != nil {
				// 	log.Fatalf("dberr %v\n", err)
				// }
				err = txn.Commit()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}

			case x := <-vmLatencyReadInfoChan:
				// fmt.Printf("-> [%v]\n", x)
				// If exist
				if _, ok := vmInfos[x.Name]; ok {
					tmp := vmInfos[x.Name]
					if x.Latency != 0 {
						tmp.LatencyRead = x.Latency
					}
					tmp.ISSSD = x.ISSSD
				} else {
					vmInfos[x.Name] = VMInformation{
						Name:        x.Name,
						LatencyRead: x.Latency,
						ISSSD:       x.ISSSD,
					}
				}
				// fmt.Printf("--> [%v]\n", vmInfos[x.Name])
				txn, err := db.Begin()
				if err != nil {
					log.Fatalf("dberr: %v\n", err)
				}
				defer txn.Rollback()
				stmt, err := txn.Prepare(`update tiramisu_state set latency=$1 where vm_name=$2`)
				if err != nil {
					log.Fatalf("dberror: %v\n", err)
				}
				// defer stmt.Close()
				if vmInfos[x.Name].LatencyRead != 0 {
					_, err = stmt.Exec(float64(vmInfos[x.Name].LatencyRead), x.Name)
				}
				// fmt.Printf("[[%v]]\n", res)
				if err != nil {
					panic(err)
				}
				// err = stmt.Close()
				// if err != nil {
				// 	log.Fatalf("dberr %v\n", err)
				// }
				err = txn.Commit()
				if err != nil {
					log.Fatalf("dberr %v\n", err)
				}
			case x := <-vmLatencyWriteInfoChan:
				// If exist
				if _, ok := vmInfos[x.Name]; ok {
					tmp := vmInfos[x.Name]
					tmp.LatencyWrite = x.Latency
					tmp.ISSSD = x.ISSSD
				} else {
					vmInfos[x.Name] = VMInformation{
						Name:         x.Name,
						LatencyWrite: x.Latency,
						ISSSD:        x.ISSSD,
					}
				}
				// fmt.Printf("--> [%v]\n", vmInfos[x.Name])
			case x := <-latencyReadChan:
				fmt.Printf("latency read c = %v v = %v\n", x.Count, x.Value)
			case x := <-latencyWriteChan:
				fmt.Printf("latency write c = %v v = %v\n", x.Count, x.Value)
			}
		}
		// After Select
		for _, e := range vmInfos {

			txn, err := db.Begin()
			if err != nil {
				log.Fatalf("dberr: %v\n", err)
			}
			defer txn.Rollback()
			stmt, err := txn.Prepare(`update tiramisu_state set latency=$1 latency_hdd=$2 latency_ssd=$3 where vm_name=$4`)
			if err != nil {
				log.Fatalf("dberror: %v\n", err)
			}
			// defer stmt.Close()
			approx_value := 0
			if e.ISSSD {
				approx_value = e.LatencyHDD
			} else {
				approx_value = e.LatencySSD
			}
			if e.ISSSD {
				_, err := stmt.Exec(e.Latency, approx_value, e.Latency, e.Name)
				//fmt.Printf("[[%v]]\n", res)
				if err != nil {
					panic(err)
				}
			} else {
				_, err := stmt.Exec(e.Latency, approx_value, e.Latency, e.Name)
				//fmt.Printf("[[%v]]\n", res)
				if err != nil {
					panic(err)
				}
			}

			// err = stmt.Close()
			// if err != nil {
			// 	log.Fatalf("dberr %v\n", err)
			// }
			err = txn.Commit()
			if err != nil {
				log.Fatalf("dberr %v\n", err)
			}
		}
	}()
	<-wait
}
