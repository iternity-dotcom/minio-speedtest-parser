package speedtest

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Bytes int64

func (b Bytes) KiB() float64 {
	return float64(b) / 1024
}

func (b Bytes) MiB() float64 {
	return b.KiB() / 1024
}

func (b Bytes) GiB() float64 {
	return b.MiB() / 1024
}

type Cluster struct {
	Info Info `json:"info"`
}

type Info struct {
	Version string `json:"minio_version"`
}

type Result struct {
	Version         string
	Network         Network         `json:"network,omitempty"`
	Drive           Drive           `json:"drive,omitempty"`
	Object          Object          `json:"object,omitempty"`
	Client          Client          `json:"client,omitempty"`
	SiteReplication SiteReplication `json:"siteReplication,omitempty"`
}

type Network struct {
	Servers []NetworkServers `json:"servers,omitempty"`
}

func (n *Network) IsPresent() bool {
	return len(n.Servers) > 0
}

type NetworkServers struct {
	Endpoint string `json:"endpoint,omitempty"`
	Nic      Perf   `json:"perf,omitempty"`
}

type Drive struct {
	Servers []DriveServers `json:"servers,omitempty"`
}

func (d *Drive) IsPresent() bool {
	return len(d.Servers) > 0
}

type DriveServers struct {
	Endpoint string `json:"endpoint,omitempty"`
	Disks    []Perf `json:"perf,omitempty"`
}

type Object struct {
	ObjectSize Bytes `json:"objectSize,omitempty"`
	Threads    int   `json:"threads,omitempty"`
	Put        Put   `json:"PUT,omitempty"`
	Get        Get   `json:"GET,omitempty"`
}

type SiteReplication struct {
	Servers []SiteReplicationServers `json:"servers,omitempty"`
}

func (s *SiteReplication) IsPresent() bool {
	return len(s.Servers) > 0
}

type SiteReplicationServers struct {
	Endpoint string `json:"endpoint,omitempty"`
	Perf     Perf   `json:"perf,omitempty"`
}

func (o *Object) IsPresent() bool {
	return o.Threads > 0
}

type Put struct {
	Perf    Perf            `json:"perf,omitempty"`
	Servers []ObjectServers `json:"servers,omitempty"`
}

type Get struct {
	Perf    Perf            `json:"perf,omitempty"`
	Servers []ObjectServers `json:"servers,omitempty"`
}

type ObjectServers struct {
	Endpoint string `json:"endpoint,omitempty"`
}

type Client struct {
	Endpoint  string        `json:"endpoint,omitempty"`
	BytesSent Bytes         `json:"bytesSent,omitempty"`
	TimeSpent time.Duration `json:"timeSpent,omitempty"`
}

func (c *Client) IsPresent() bool {
	return c.Endpoint != ""
}

func (c Client) Throughput() Bytes {
	b := int64(c.BytesSent)
	s := int64(c.TimeSpent.Seconds())
	return Bytes(b / s)
}

type Perf struct {
	Throughput      Bytes         `json:"throughput,omitempty"`
	ObjectsPerSec   int           `json:"objectsPerSec,omitempty"`
	ResponseTime    ResponseTime  `json:"responseTime,omitempty"`
	Ttfb            ResponseTime  `json:"ttfb,omitempty"`
	Tx              Bytes         `json:"tx,omitempty"`
	TxTotalDuration time.Duration `json:"txTotalDuration,omitempty"`
	Rx              Bytes         `json:"rx,omitempty"`
	RxTotalDuration time.Duration `json:"rxTotalDuration,omitempty"`
	Path            string        `json:"path,omitempty"`
	ReadThroughput  Bytes         `json:"readThroughput,omitempty"`
	WriteThroughput Bytes         `json:"writeThroughput,omitempty"`
}

func (p *Perf) RxBps() Bytes {
	if p.RxTotalDuration > 0 {
		return Bytes(int64(p.Rx) / int64(p.RxTotalDuration.Seconds()))
	}
	return p.Rx
}

func (p *Perf) TxBps() Bytes {
	if p.TxTotalDuration > 0 {
		return Bytes(int64(p.Tx) / int64(p.TxTotalDuration.Seconds()))
	}
	return p.Tx
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

	if r.Network.IsPresent() {
		fmt.Fprintln(&buf, "NetPerf: ✔")
		fmt.Fprintln(&buf, "")
		fmt.Fprintln(&buf, "NODE\t\t\t\tRX\t\tTX")
		for _, s := range r.Network.Servers {
			fmt.Fprintf(&buf, "%s\t%.1f GiB/s\t%.1f GiB/s\n", s.Endpoint, s.Nic.RxBps().GiB(), s.Nic.TxBps().GiB())
		}
		fmt.Fprintln(&buf, "")
	}

	if r.Drive.IsPresent() {
		fmt.Fprintln(&buf, "DrivePerf: ✔")
		fmt.Fprintln(&buf, "")
		fmt.Fprintln(&buf, "NODE\t\t\t\tPATH\t\t\tREAD\t\tWRITE")
		for _, server := range r.Drive.Servers {
			for _, disk := range server.Disks {
				driveCount++
				fmt.Fprintf(&buf, "%s\t%s\t%.0f MiB/s\t%.0f MiB/s\n", server.Endpoint, disk.Path, disk.ReadThroughput.MiB(), disk.WriteThroughput.MiB())
			}
		}
		fmt.Fprintln(&buf, "")
	}

	if r.Object.IsPresent() {
		fmt.Fprintln(&buf, "ObjectPerf: ✔")
		fmt.Fprintln(&buf, "")
		fmt.Fprintln(&buf, "   \tTHROUGHPUT\tIOPS")
		fmt.Fprintf(&buf, "PUT\t%.1f GiB/s\t%d objs/s\n", r.Object.Put.Perf.Throughput.GiB(), r.Object.Put.Perf.ObjectsPerSec)
		fmt.Fprintf(&buf, "GET\t%.1f GiB/s\t%d objs/s\n", r.Object.Get.Perf.Throughput.GiB(), r.Object.Get.Perf.ObjectsPerSec)
		fmt.Fprintln(&buf, "")
		fmt.Fprintf(&buf, "MinIO %s, %d servers, %d drives, %.0f MiB objects, %d threads\n", r.Version, len(r.Network.Servers), driveCount, r.Object.ObjectSize.MiB(), r.Object.Threads)
		fmt.Fprintln(&buf, "")
	}

	if r.Client.IsPresent() {
		fmt.Fprintln(&buf, "Client: ✔")
		fmt.Fprintln(&buf, "")
		fmt.Fprintln(&buf, "ENDPOINT\t\t\t\t\tTX")
		fmt.Fprintf(&buf, "%s\t%.1f MiB/s\n", r.Client.Endpoint, r.Client.Throughput().MiB())
	}

	if r.SiteReplication.IsPresent() {
		fmt.Fprintln(&buf, "SiteReplication: ✔")
		fmt.Fprintln(&buf, "")
		fmt.Fprintln(&buf, "ENDPOINT\t\tRX\t\tTX")
		for _, s := range r.SiteReplication.Servers {
			fmt.Fprintf(&buf, "%s\t%.1f MiB/s\t%.1f MiB/s\n", s.Endpoint, s.Perf.RxBps().MiB(), s.Perf.TxBps().MiB())
		}
		fmt.Fprintln(&buf, "")
	}

	return buf.String()
}

func FromZipFile(zipFile string) (*Result, error) {

	zip, err := zip.OpenReader(zipFile)
	if err != nil {
		return nil, err
	}
	defer zip.Close()
	var result *Result
	var cluster *Cluster
	for _, f := range zip.File {
		srcFile, err := f.Open()
		if err != nil {
			continue
		}
		defer srcFile.Close()
		b, err := io.ReadAll(srcFile)

		if err != nil {
			continue
		}
		if f.Name == "cluster.info" {
			cluster, _ = clusterFromJsonByteArray(b)
		}
		if strings.HasSuffix(f.Name, "json") {
			result, _ = FromJsonByteArray(b)
		}

	}
	if result != nil {
		result.Version = cluster.Info.Version
		return result, nil
	}
	return nil, fmt.Errorf("no valid .json file in .zip archive %s", zipFile)
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

func clusterFromJsonByteArray(jsonData []byte) (*Cluster, error) {
	b := &Cluster{}
	err := json.Unmarshal(jsonData, b)
	return b, err
}
