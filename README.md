# yuscan
 东抄西凑的go的扫描器



功能：

- [ ] 端口扫描
- [ ] 子域名枚举
- [ ] 扫描代理; 替代方式：kali proxychain
- [ ] 端口扫描走代理存在问题，走代理后，每一个端口都显示open。走代理后，扫描器和代理之间建立的连接被认为成功建立连接，然后直接返回conn，显示open。我没想到扫描好的解决办法。因为是抄fscan的，fscan的socks5代理也有问题，端口扫描不走httpproxy。

编译:
go build

命令：
端口扫描
yuscan.exe -model  portscan -hostslist hostslist.txt -ports 1-1024,7000-8000,6379
    端口扫描，默认是会加入常用端口的，后续会加入一个参数，如果存在就不加入

   - [ ] 检测活动端口是什么端口，比如22-->ssh
   - [ ] 如果判断是http/https，接着接入指纹扫描，判断框架

子域名扫描

yuscan.exe -domain baidu.com -wordlist namelist.txt

   - [x] 加入对泛解析的探测 https://mp.weixin.qq.com/s/MfA-lkYIRNJtSTNTZDqmng

参考项目：
fscan   直接抄的，源码都没改。
bhg github.com/blackhat-go/bhg 也是抄的，源码都没改

模块解读:
yuscan/common
存放常用的方法、配置等信息
功能介绍:
func ParsePort(ports string) (scanPorts []int)
    端口的string ==> 字符串数组
    "1-3,15,15,16-17" ==> [1,2,3,15,16,17]

func RandAllString(lenNum int) string
func RandNumString(lenNum int) string
func RandString(lenNum int) string
    生成随机字符串