package main

import "time"

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

type VMInformation struct {
	Name         string
	Latency      int
	LatencyRead  int
	LatencyWrite int
	IOPS         int
	LatencyHDD   int
	IOPSHDD      int
	LatencySSD   int
	IOPSSSD      int
	ISSSD        bool
	LastUpdated  int64
}

type VMstate struct {
	Name        string  `gorm:"column:vm_name;primary_key"`
	Latency     float64 `gorm:"column:latency"`
	IOPS        float64 `gorm:"column:iops"`
	Latency_HDD float64 `gorm:"column:latency_hdd"`
	IOPS_HDD    float64 `gorm:"column:iops_hdd"`
	Latency_SSD float64 `gorm:"column:latency_ssd"`
	IOPS_SSD    float64 `gorm:"column:iops_ssd"`
	CreatedAt   time.Time
	UpdateAt    time.Time
	DeleteAt    *time.Time
}

func (VMstate) TableName() string {
	return "tiramisu_state"
}
