/*************************************************************************
   > File Name: nginx_process.go
   > Author: Kee
   > Mail: chinboy2012@gmail.com
   > Created Time: 2017.12.06
************************************************************************/
package nginx

import (
	//"errors"
	"fmt"
	"github.com/keesely/kfiles"
	"strconv"
	"strings"
)

// 获取 Nginx 进程PID
func getPid(this *Nginx) (int, error) {
	if exists := kfiles.Exists(this.Pid); exists == false {
		return 0, nil
	}

	fPid, err := kfiles.Get(this.Pid)

	if err != nil {
		return 0, err
	}

	sPid := strings.Replace(fPid, "\n", "", -1)
	sPid = fmt.Sprintf("%s", sPid)
	pid, err := strconv.Atoi(sPid)

	return pid, err
}
