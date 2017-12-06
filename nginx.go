/*************************************************************************
   > File Name: nginx.go
   > Author: Kee
   > Mail: chinboy2012@gmail.com
   > Created Time: 2017.12.04
************************************************************************/
package nginx

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/process"
	"os/exec"
	"strings"
	"syscall"
	"time"
	//"strconv"
	//"github.com/keesely/kfiles"
	//"github.com/shirou/gopsutil/internal/common"
)

type Nginx struct {
	Pid    string // Nginx PID文件
	Nginx  string // Nginx 可执行文件
	prcess *process.Process
}

type Memory struct {
	Percent     float32 `json:"percent"`
	VirtualSize uint64  `json:"virtual_size"`
	RealSize    uint64  `json:"real_size"`
}

type Status struct {
	PID      int32   `json:"pid"`
	CPU      float32 `json:"cpu"`
	Memory   *Memory `json:"memory"`
	Status   string  `json:"status"`
	Start    string  `json:"start_at"`
	Time     string  `json:"time"`
	Hostname string  `json:"hostname"`
	Subpid   []int32 `json:"sub_pid"`
}

type Result struct {
	Code int32        `json:"code"`
	Msg  string       `json:"msg"`
	Data *interface{} `json:"data"`
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

/**
 * 获取 Nginx 状态
 *
 * @return *Status, error
 */
func (this *Nginx) Status() (*Status, error) {
	pid, err := getPid(this)

	if err != nil {
		return nil, err
	}

	if pexis, _ := process.PidExists(int32(pid)); pexis == false {
		return nil, errors.New("进程ID不存在")
	}

	proc, err := process.NewProcess(int32(pid))

	// CPU占用比例
	cpu, _ := proc.CPUPercent()

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

	// 运行状态
	status, _ := proc.Status()

	// 启动时间
	start_f, _ := proc.CreateTime()
	start := time.Unix(start_f/1000, start_f).Format(time.RFC3339)

	// 运行时长
	ttl_f := time.Now().Sub(time.Unix(start_f/1000, start_f)).Seconds()
	ttl := fmt.Sprintf("%.5f", ttl_f)

	// 主机名
	host, _ := host.Info()

	// 子进程
	children, serr := proc.Children()

	sub := make([]int32, 0)

	if serr == nil {
		for _, spid := range children {
			sub = append(sub, int32(spid.Pid))
		}
	}

	p := &Status{
		PID:      proc.Pid,
		CPU:      float32(cpu),
		Memory:   mem,
		Status:   status,
		Start:    start,
		Time:     ttl,
		Hostname: host.Hostname,
		Subpid:   sub,
	}

	return p, err
}

// 启动 Nginx 服务
func (this *Nginx) Start() (bool, error) {
	status, _ := this.Status()

	if status != nil {
		return true, errors.New("Nginx服务已经启动")
	}

	// 测试文档
	test, terr := this.Test()
	if false == test {
		return false, terr
	}

	start := exec.Command("/bin/sh", "-c", this.Nginx)
	startResult, err := start.CombinedOutput()

	if string(startResult) != "" {
		return false, errors.New("nginx starting Result: \n" + string(startResult))
	}
	if err != nil {
		return false, errors.New("Start error: " + err.Error())
	}

	return true, nil
}

// 重载 Nginx 服务
func (this *Nginx) Reload() (bool, error) {
	status, _ := this.Status()

	if status == nil {
		return this.Start()
	}

	test, terr := this.Test()
	if false == test {
		return false, terr
	}

	pid := status.PID
	err := syscall.Kill(int(pid), syscall.SIGHUP)
	if err != nil {
		return false, err
	}

	return true, nil
}

// 停止 Nginx 服务
func (this *Nginx) Stop() (bool, error) {
	status, _ := this.Status()

	if 0 >= status.PID {
		return true, nil
	}

	stop := exec.Command("/bin/sh", "-c", this.Nginx+" -s stop")
	stopResult, err := stop.CombinedOutput()
	if string(stopResult) != "" {
		return false, errors.New("nginx stopped Result:\n" + string(stopResult))
	}

	if err != nil {
		return false, errors.New("nginx -s stop Error: \n" + (err.Error()))
	}

	return true, nil
}

// 测试 Nginx 配置
func (this *Nginx) Test() (bool, error) {
	test := exec.Command("/bin/sh", "-c", this.Nginx+" -t")

	testResult, err := test.CombinedOutput()

	if err != nil {
		return false, err
	}

	result := "nginx testing result:\n" + string(testResult)

	if !strings.Contains(string(testResult), "successful") {
		return false, errors.New(result)
	}

	return true, nil
}
