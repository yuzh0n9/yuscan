package Plugins

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"yuscan/common"
)

type Addr struct {
	ip   string
	port int
}

/*
作用：端口扫描
hostslist 主机名 []string
ports 需要扫描的端口号 string
timeout 超时时间 int64

返回值：AliveAddress 活动地址 []string
*/
func PortScan(hostslistFile *string, ports string, timeout int64) []string {
	var AliveAddress []string
	var hostslist []string

	// 字符串拼接，加入默认的端口
	var build strings.Builder
	build.WriteString(common.DefaultPorts)
	build.WriteString(",")
	build.WriteString(common.Webport)
	build.WriteString(",")
	build.WriteString(ports)
	ports = build.String()
	probePorts := common.ParsePort(ports)
	noPorts := common.ParsePort(common.NoPorts)

	// 创建一个新的scanner
	fh, err := os.Open(*hostslistFile)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	scanner := bufio.NewScanner(fh) //按行读文件
	for scanner.Scan() {
		hostslist = append(hostslist, scanner.Text())
	}

	// noPorts 没看懂这一块
	if len(noPorts) > 0 {
		temp := map[int]struct{}{}
		for _, port := range probePorts {
			temp[port] = struct{}{}
		}

		for _, port := range noPorts {
			delete(temp, port)
		}

		var newDatas []int
		for port, _ := range temp {
			newDatas = append(newDatas, port)
		}
		probePorts = newDatas
		sort.Ints(probePorts)
	}
	workers := common.Threads                                    //几个线程
	Addrs := make(chan Addr, len(hostslist)*len(probePorts))     // 管道
	results := make(chan string, len(hostslist)*len(probePorts)) // 管道，结果
	var wg sync.WaitGroup

	//接收结果
	go func() {
		for found := range results {
			AliveAddress = append(AliveAddress, found)
			wg.Done()
		}
	}()

	//多线程扫描
	for i := 0; i < workers; i++ {
		go func() {
			for addr := range Addrs { // addr hostname:port
				PortConnect(addr, results, timeout, &wg)
				wg.Done()
			}
		}()
	}

	//添加扫描目标
	for _, port := range probePorts {
		for _, host := range hostslist {
			wg.Add(1)
			Addrs <- Addr{host, port}
		}
	}
	wg.Wait()
	close(Addrs)
	close(results)
	return AliveAddress
}

// 端口探测
func PortConnect(addr Addr, respondingHosts chan<- string, adjustedTimeout int64, wg *sync.WaitGroup) {
	host, port := addr.ip, addr.port
	// tcp4 ==> ipv4 only; tcp ==> default ipv4
	conn, err := common.WrapperTcpWithTimeout("tcp4", fmt.Sprintf("%s:%v", host, port), time.Duration(adjustedTimeout)*time.Second)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if err == nil {
		address := host + ":" + strconv.Itoa(port)
		result := fmt.Sprintf("%s open", address)
		common.LogSuccess(result)
		wg.Add(1)
		respondingHosts <- address
	}
}

func NoPortScan(hostslist []string, ports string) (AliveAddress []string) {
	probePorts := common.ParsePort(ports)
	noPorts := common.ParsePort(common.NoPorts)
	if len(noPorts) > 0 {
		temp := map[int]struct{}{}
		for _, port := range probePorts {
			temp[port] = struct{}{}
		}

		for _, port := range noPorts {
			delete(temp, port)
		}

		var newDatas []int
		for port, _ := range temp {
			newDatas = append(newDatas, port)
		}
		probePorts = newDatas
		sort.Ints(probePorts)
	}
	for _, port := range probePorts {
		for _, host := range hostslist {
			address := host + ":" + strconv.Itoa(port)
			AliveAddress = append(AliveAddress, address)
		}
	}
	return
}
