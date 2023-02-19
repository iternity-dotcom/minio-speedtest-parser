package speedtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

type Bytes int64

func (b Bytes) KiB() float64 {
	return float64(b) / 1000
}

func (b Bytes) MiB() float64 {
	return b.KiB() / 1000
}

func (b Bytes) GiB() float64 {
	return b.MiB() / 1000
}

type Result struct {
	Object  Object  `json:"object,omitempty"`
	Network Network `json:"network,omitempty"`
	Drive   Drive   `json:"drive,omitempty"`
}

type Object struct {
	ObjectSize Bytes `json:"objectSize,omitempty"`
	Threads    int   `json:"threads,omitempty"`
	Put        Put   `json:"PUT,omitempty"`
	Get        Get   `json:"GET,omitempty"`
}

type Network struct {
	Servers []NetworkServers `json:"servers,omitempty"`
}

type Drive struct {
	Servers []DriveServers `json:"servers,omitempty"`
}

type Put struct {
	Perf    Perf            `json:"perf,omitempty"`
	Servers []ObjectServers `json:"servers,omitempty"`
}

type Get struct {
	Perf    Perf            `json:"perf,omitempty"`
	Servers []ObjectServers `json:"servers,omitempty"`
}

type NetworkServers struct {
	Endpoint string `json:"endpoint,omitempty"`
	Nic      Perf   `json:"perf,omitempty"`
}

type DriveServers struct {
	Endpoint string `json:"endpoint,omitempty"`
	Disks    []Perf `json:"perf,omitempty"`
}

type ObjectServers struct {
	Endpoint string `json:"endpoint,omitempty"`
}

type Perf struct {
	Throughput      Bytes        `json:"throughput,omitempty"`
	ObjectsPerSec   int          `json:"objectsPerSec,omitempty"`
	ResponseTime    ResponseTime `json:"responseTime,omitempty"`
	Ttfb            ResponseTime `json:"ttfb,omitempty"`
	Tx              Bytes        `json:"tx,omitempty"`
	Rx              Bytes        `json:"rx,omitempty"`
	Path            string       `json:"path,omitempty"`
	ReadThroughput  Bytes        `json:"readThroughput,omitempty"`
	WriteThroughput Bytes        `json:"writeThroughput,omitempty"`
}

type ResponseTime struct {
	Avg   Bytes `json:"avg,omitempty"`
	P50   Bytes `json:"p50,omitempty"`
	P75   Bytes `json:"p75,omitempty"`
	P95   Bytes `json:"p95,omitempty"`
	P99   Bytes `json:"p99,omitempty"`
	P999  Bytes `json:"p999,omitempty"`
	L5P   Bytes `json:"l5p,omitempty"`
	S5P   Bytes `json:"s5p,omitempty"`
	Max   Bytes `json:"max,omitempty"`
	Min   Bytes `json:"min,omitempty"`
	Sdev  Bytes `json:"sdev,omitempty"`
	Range Bytes `json:"range,omitempty"`
}

func (r *Result) ToJson() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

func (r *Result) String() string {
	var buf bytes.Buffer
	driveCount := 0

	fmt.Fprintln(&buf, "NetPerf: ✔")
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, "NODE\t\t\t\t\tRX\t\tTX")
	for _, s := range r.Network.Servers {
		fmt.Fprintf(&buf, "%s\t%.1f GiB/s\t%.1f GiB/s\n", s.Endpoint, s.Nic.Rx.GiB(), s.Nic.Tx.GiB())
	}
	fmt.Fprintln(&buf, "")

	fmt.Fprintln(&buf, "DrivePerf: ✔")
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, "NODE\t\t\t\t\tPATH\t\t\tREAD\t\tWRITE")
	for _, server := range r.Drive.Servers {
		for _, disk := range server.Disks {
			driveCount++
			fmt.Fprintf(&buf, "%s\t%s\t%.0f MiB/s\t%.0f MiB/s\n", server.Endpoint, disk.Path, disk.ReadThroughput.MiB(), disk.WriteThroughput.MiB())
		}
	}
	fmt.Fprintln(&buf, "")

	fmt.Fprintln(&buf, "ObjectPerf: ✔")
	fmt.Fprintln(&buf, "")
	fmt.Fprintln(&buf, "   \tTHROUGHPUT\tIOPS")
	fmt.Fprintf(&buf, "PUT\t%.1f GiB/s\t%d objs/s\n", r.Object.Put.Perf.Throughput.GiB(), r.Object.Put.Perf.ObjectsPerSec)
	fmt.Fprintf(&buf, "GET\t%.1f GiB/s\t%d objs/s\n", r.Object.Get.Perf.Throughput.GiB(), r.Object.Get.Perf.ObjectsPerSec)
	fmt.Fprintln(&buf, "")
	fmt.Fprintf(&buf, "%d servers, %d drives, %.0f MiB objects, %d threads", len(r.Network.Servers), driveCount, r.Object.ObjectSize.MiB(), r.Object.Threads)

	return buf.String()
}

func FromJsonFile(jsonFile string) (*Result, error) {
	jsonData, err := os.ReadFile(jsonFile)
	if err != nil {
		return nil, err
	}
	return FromJsonByteArray(jsonData)
}

func FromJsonByteArray(jsonData []byte) (*Result, error) {
	b := &Result{}
	err := json.Unmarshal(jsonData, b)
	return b, err
}
