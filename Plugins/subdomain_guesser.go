package Plugins

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"yuscan/common"

	"github.com/miekg/dns"
)

// 查询A记录
// 接受FQDN作为第一个参数，并接受DNS服务器的地址作为第二个参数。函数返回一个字符串切片和一个错误。
func lookupA(fqdn, serverAddr string) ([]string, error) {
	var m dns.Msg
	var ips []string
	m.SetQuestion(dns.Fqdn(fqdn), dns.TypeA)
	in, err := dns.Exchange(&m, serverAddr)
	if err != nil {
		return ips, err
	}
	if len(in.Answer) < 1 {
		return ips, errors.New("no answer")
	}
	for _, answer := range in.Answer {
		if a, ok := answer.(*dns.A); ok {
			ips = append(ips, a.A.String())
		}
	}
	return ips, nil
}

// 查询CNAME记录
func lookupCNAME(fqdn, serverAddr string) ([]string, error) {
	var m dns.Msg
	var fqdns []string
	m.SetQuestion(dns.Fqdn(fqdn), dns.TypeCNAME)
	in, err := dns.Exchange(&m, serverAddr)
	if err != nil {
		return fqdns, err
	}
	if len(in.Answer) < 1 {
		return fqdns, errors.New("no answer")
	}
	for _, answer := range in.Answer {
		if c, ok := answer.(*dns.CNAME); ok {
			fqdns = append(fqdns, c.Target)
		}
	}
	return fqdns, nil
}

func lookup(fqdn, serverAddr string) []result {
	var results []result
	var cfqdn = fqdn // Don't modify the original.
	for {
		cnames, err := lookupCNAME(cfqdn, serverAddr)
		if err == nil && len(cnames) > 0 {
			cfqdn = cnames[0]
			continue // We have to process the next CNAME.
		}
		ips, err := lookupA(cfqdn, serverAddr)
		if err != nil {
			break // There are no A records for this hostname.
		}
		for _, ip := range ips {
			results = append(results, result{IPAddress: ip, Hostname: fqdn})
		}
		break // We have processed all the results.
	}
	return results
}

func sub_worker(tracker chan empty, fqdns chan string, gather chan []result, serverAddr string) {
	for fqdn := range fqdns {
		results := lookup(fqdn, serverAddr)
		if len(results) > 0 {
			gather <- results
		}
	}
	var e empty
	tracker <- e
}

type empty struct{}

type result struct {
	IPAddress string
	Hostname  string
}

func Subdomain_guesser(flDomain *string, flWordlist *string, flWorkerCount *int, flServerAddr *string) {

	var results []result

	fqdns := make(chan string, *flWorkerCount)
	gather := make(chan []result) // 存放结果
	tracker := make(chan empty)

	// 检测是否存在泛解析
	// 生成5个8位随机字符串，然后拼接域名，如果随机域名可以解析，说明存在泛解析
	// 泛解析的案例: xxxx.taobao.com
	var count_ran = 0
	for i := 0; i < 5; i++ {
		ran_domain := fmt.Sprintf("%s.%s", common.RandAllString(8), *flDomain)
		// fmt.Println(ran_domain)
		result := lookup(ran_domain, *flServerAddr)
		if result != nil {
			count_ran++
		}
	}
	if count_ran > 3 {
		fmt.Println("该域名存在泛解析")
		os.Exit(1)
	}

	// 创建一个新的scanner
	fh, err := os.Open(*flWordlist)
	if err != nil {
		panic(err)
	}
	defer fh.Close()
	scanner := bufio.NewScanner(fh) //按行读文件

	// 启动工人函数
	for i := 0; i < *flWorkerCount; i++ {
		go sub_worker(tracker, fqdns, gather, *flServerAddr)
	}

	go func() {
		for r := range gather {
			results = append(results, r...)
		}
		var e empty
		tracker <- e
	}()

	for scanner.Scan() {
		fqdns <- fmt.Sprintf("%s.%s", scanner.Text(), *flDomain)
	}
	// Note: We could check scanner.Err() here.

	close(fqdns)
	for i := 0; i < *flWorkerCount; i++ {
		<-tracker
	}
	close(gather)
	<-tracker

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', 0)
	for _, r := range results {
		fmt.Fprintf(w, "%s\t%s\n", r.Hostname, r.IPAddress)
		common.LogSuccess(fmt.Sprintf("%s  %s\n", r.Hostname, r.IPAddress))
	}
	w.Flush()
}
