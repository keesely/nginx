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
	pnet "github.com/shirou/gopsutil/net"
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
	PID      int32          `json:"pid"`
	CPU      float32        `json:"cpu"`
	Memory   *Memory        `json:"memory"`
	Status   string         `json:"status"`
	Start    string         `json:"start_at"`
	Time     float32        `json:"time"`
	Host     *host.InfoStat `json:"host"`
	Subpid   []int32        `json:"sub_pid"`
	IpAddrs  []string       `json:"ip_address"`
	Networks *Networks      `json:"networks"`
}

type Process struct {
	Pid     int32
	Process *process.Process
}

type Addr struct {
	IP   string `json:"ip"`
	Port uint32 `json:port`
}

type Network struct {
	Stat  string `json:"stat"`
	Laddr string `json:"local_address"`
	Raddr string `json:"remote_address"`
}

type Networks struct {
	Network    []*Network            `json:"network"`
	IOCounters []pnet.IOCountersStat `json:"io_counters"`
	Total      map[string]int        `json:"total"`
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

// 启动时间格式化
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
	subs := make([]int32, 0)
	pid := this.Process.Pid

	procs, err := process.Processes()
	if err != err {
		return subs
	}

	for _, sub := range procs {
		if ppid, _ := sub.Ppid(); ppid == pid {
			subs = append(subs, int32(sub.Pid))
		}
	}

	return subs
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

// 获取网络情况
func (this *Process) Networks() *Networks {
	pid := this.Process.Pid

	status_list := map[string]int{"LISTEN": 0, "ESTABLISHED": 0, "TIME_WAIT": 0, "CLOSE_WAIT": 0, "LAST_ACK": 0, "SYN_SENT": 0}
	//status_list := &NetTotal{LISTEN: 0, ESTABLISHED: 0, TIME_WAIT: 0, CLOSE_WAIT: 0, LAST_ACK: 0, SYN_SENT: 0}

	/**
	networks := &Networks{
		LISTEN:      make([]*Network, 0),
		ESTABLISHED: make([]*Network, 0),
		TIME_WAIT:   make([]*Network, 0),
		CLOSE_WAIT:  make([]*Network, 0),
		LAST_ACK:    make([]*Network, 0),
		SYN_SENT:    make([]*Network, 0),
		Total:       status_list,
	}
	*/
	networks := make([]*Network, 0)

	pn, _ := pnet.ConnectionsPid("tcp", int32(pid))

	for _, sub_pc := range pn {
		net := &Network{
			Stat:  sub_pc.Status,
			Laddr: sub_pc.Laddr.IP + ":" + fmt.Sprintf("%d", sub_pc.Laddr.Port),
			Raddr: sub_pc.Raddr.IP + ":" + fmt.Sprintf("%d", sub_pc.Raddr.Port),
			//Laddr: &Addr{IP: sub_pc.Laddr.IP, Port: sub_pc.Laddr.Port},
			//Raddr: &Addr{IP: sub_pc.Raddr.IP, Port: sub_pc.Raddr.Port},
		}

		status := string(sub_pc.Status)

		networks = append(networks, net)

		status_list[status] += 1
	}

	n, _ := this.Process.NetIOCounters(false)

	return &Networks{
		Network:    networks,
		Total:      status_list,
		IOCounters: n,
	}
}

// 获取 Nginx 进程PID
func getPid(this *Nginx) (int32, error) {
	if exists := kfiles.Exists(this.Pid); exists == false {
		return int32(0), errors.New("PID文件不存在 : " + this.Pid)
	}

	fPid, err := kfiles.Get(this.Pid)

	if err != nil {
		return int32(0), err
	}

	sPid := strings.Replace(fPid, "\n", "", -1)
	sPid = fmt.Sprintf("%s", sPid)
	if sPid == "" {
		return int32(0), errors.New("PID不存在 : " + this.Pid)
	}
	pid, err := strconv.Atoi(sPid)

	return int32(pid), err
}
