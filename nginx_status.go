/*************************************************************************
   > File Name: nginx_status.go
   > Author: Kee
   > Mail: chinboy2012@gmail.com
   > Created Time: 2017.12.06
************************************************************************/
package nginx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keesely/kfiles"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
	"net"
	"strconv"
	"strings"
	"time"
)

type Memory struct {
	Percent     float32 `json:"percent"`
	VirtualSize uint64  `json:"virtual_byte"`
	RealSize    uint64  `json:"real_byte"`
}

type Status struct {
	PID     int32          `json:"pid"`
	CPU     float32        `json:"cpu"`
	Memory  *Memory        `json:"memory"`
	Status  string         `json:"status"`
	Start   string         `json:"start_at"`
	Time    float32        `json:"time"`
	Host    *host.InfoStat `json:"host"`
	IpAddrs []string       `json:"ip_address"`
	Subpid  []int32        `json:"sub_pid"`
}

type Process struct {
	Pid     int32
	Process *process.Process
}

// String returns JSON value of the memory info
func (obj *Memory) String() string {
	str, _ := json.Marshal(obj)
	return string(str)
}

// String returns JSON value of the status info
func (obj *Status) String() string {
	str, _ := json.Marshal(obj)
	return string(str)
}

func (this *Process) New(pid int32) (*Process, error) {
	if pexis, _ := process.PidExists(int32(pid)); pexis == false {
		return nil, errors.New("进程ID不存在")
	}

	proc, err := process.NewProcess(int32(pid))

	if err != nil {
		return nil, err
	}
	self := &Process{
		Pid:     proc.Pid,
		Process: proc,
	}
	return self, nil
}

// CPU占用比例
func (this *Process) Cpu() float64 {
	proc := this.Process

	cpu, _ := proc.CPUPercent()

	return cpu
}

// 内存信息
func (this *Process) Memory() *Memory {
	proc := this.Process

	// 内存信息
	mem_info, _ := proc.MemoryInfo()
	// 内存占用比例
	m_percent, _ := proc.MemoryPercent()

	// 内存详情
	mem := &Memory{
		Percent:     m_percent,
		VirtualSize: mem_info.VMS,
		RealSize:    mem_info.RSS,
	}

	return mem
}

// 运行状态
func (this *Process) Status() string {
	status, _ := this.Process.Status()
	return status
}

// 启动时间
func (this *Process) CreateTime() int64 {
	start_f, _ := this.Process.CreateTime()

	return start_f
}

func (this *Process) StartDateTime() string {
	start_f := this.CreateTime()
	start := time.Unix(start_f/1000, start_f).Format(time.RFC3339)
	return start
}

// 运行时长
func (this *Process) Time() float32 {
	start_f := this.CreateTime()

	ttl_f := time.Now().Sub(time.Unix(start_f/1000, start_f)).Seconds()
	//ttl := fmt.Sprintf("%.5f", ttl_f)

	return float32(ttl_f)
}

// 主机名
func (this *Process) Host() *host.InfoStat {
	host, _ := host.Info()
	return host
}

// 子进程 PID 列表
func (this *Process) Children() []int32 {
	children, err := this.Process.Children()

	sub := make([]int32, 0)

	if err == nil {
		for _, spid := range children {
			sub = append(sub, int32(spid.Pid))
		}
	}

	return sub
}

// 获取网卡IP
func (this *Process) Internal() []string {
	addrs, err := net.InterfaceAddrs()

	ip := make([]string, 0)

	if err != nil {
		return ip
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = append(ip, ipnet.IP.String())
			}
		}
	}

	return ip
}

// 获取 Nginx 进程PID
func getPid(this *Nginx) (int32, error) {
	if exists := kfiles.Exists(this.Pid); exists == false {
		return int32(0), errors.New("PID文件不存在")
	}

	fPid, err := kfiles.Get(this.Pid)

	if err != nil {
		return int32(0), err
	}

	sPid := strings.Replace(fPid, "\n", "", -1)
	sPid = fmt.Sprintf("%s", sPid)
	if sPid == "" {
		return int32(0), errors.New("PID不存在")
	}
	pid, err := strconv.Atoi(sPid)

	return int32(pid), err
}
