/*************************************************************************
   > File Name: nginx.go
   > Author: Kee
   > Mail: chinboy2012@gmail.com
   > Created Time: 2017.12.04
************************************************************************/
package nginx

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	//"strconv"
	//"github.com/keesely/kfiles"
	//"github.com/shirou/gopsutil/internal/common"
)

type Nginx struct {
	Pid   string // Nginx PID文件
	Nginx string // Nginx 可执行文件
}

type Result struct {
	Code int32        `json:"code"`
	Msg  string       `json:"msg"`
	Data *interface{} `json:"data"`
}

/**
 * 获取 Nginx 状态
 *
 * @return *Status, error
 */
func (this *Nginx) Status() (*Status, error) {
	pid, err := getPid(this)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	process := new(Process)
	proc, perr := process.New(int32(pid))

	if perr != nil {
		return nil, perr
	}

	p := &Status{
		PID:     proc.Pid,
		CPU:     float32(proc.Cpu()),
		Memory:  proc.Memory(),
		Status:  proc.Status(),
		Start:   proc.StartDateTime(),
		Time:    proc.Time(),
		Host:    proc.Host(),
		IpAddrs: proc.Internal(),
		Subpid:  proc.Children(),
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
