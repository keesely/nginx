/*************************************************************************
   > File Name: nginx_test.go
   > Author: Kee
   > Mail: chinboy2012@gmail.com
   > Created Time: 2017.12.05
************************************************************************/
package nginx

import (
	"fmt"
	"testing"
)

func getNgx() *Nginx {
	ngx := new(Nginx)

	ngx.Pid = "/home/nginx/logs/nginx.pid"
	ngx.Nginx = "/home/nginx/sbin/nginx"

	return ngx
}

func Test(t *testing.T) {

	start, err2 := getNgx().Start()
	if start == false {
		fmt.Println(err2)
	} else {
		fmt.Println("Nginx 服务启动成功")
	}

	status, err := getNgx().Status()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(status.String())

		// 执行重载
		reload, er2 := getNgx().Reload()

		if false == reload {
			fmt.Println(er2)
		} else {
			status, err = getNgx().Status()

			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(status.String())
			}
		}
	}
}

func StopTest(t *testing.T) {
	stop, err := getNgx().Stop()
	if stop == false {
		fmt.Println(err)
	} else {
		fmt.Println("Nginx 服务已经停止")
	}
}
