package main

import (
	"flag"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	netIp           string // ip
	configPath      string //配置文件地址
	accessKey       string // 阿里云accessKey
	accessKeySecret string // 阿里云accessKeySecret
	regionID        string // 阿里云regionID
	domain          string // 需要绑定的域名
	rrs             string // 需要子域
)

// 指定配置路径
func init() {
	flag.StringVar(&configPath, "c", "conf.ini", "conf.ini path")
}

// 获取外网IP
func get_external() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	return string(content)
}

// 阿里云ddns绑定
func replace(ip string) {
	client, err := alidns.NewClientWithAccessKey(
		regionID,      // 您的可用区ID
		accessKey, // 您的AccessKey ID
		accessKeySecret) // 您的AccessKey Secret
	if err != nil {
		// 异常处理
		panic(err)
	}
	upfunc := func(rr string) {
		request := alidns.CreateDescribeSubDomainRecordsRequest()
		request.SubDomain = fmt.Sprintf("%s.%s", rr,domain)
		response, err := client.DescribeSubDomainRecords(request)
		if err != nil {
			// 异常处理
			log.Println(err,request)
		}
		if response.IsSuccess() {
			listRecords := response.DomainRecords.Record
			if len(listRecords) >= 1 {
				for _, record := range listRecords {
					if record.Value != ip {
						request := alidns.CreateUpdateDomainRecordRequest()
						request.Value = ip
						request.RecordId = record.RecordId
						request.RR = rr
						request.TTL = "600"
						request.Type = "A"
						if res, err := client.UpdateDomainRecord(request); err != nil {
							// 异常处理
							log.Println(err,request)
						} else {
							if !res.IsSuccess() {
								log.Println(res.String())
							}else {
								log.Printf("replace domain %s ip %s", fmt.Sprintf("%s.%s", rr,domain),ip)
							}
						}
					}
				}
			} else {
				request := alidns.CreateAddDomainRecordRequest()
				request.Value = ip
				request.RR = rr
				request.DomainName = domain
				request.TTL = "600"
				request.Type = "A"
				if res, err := client.AddDomainRecord(request); err != nil {
					// 异常处理
					log.Println(err,request)
				} else {
					if !res.IsSuccess() {
						log.Println(res.String())
					}else {
						log.Printf("replace domain %s ip %s", fmt.Sprintf("%s.%s", rr,domain),ip)
					}
				}
			}
		}
	}
	rrsList := strings.Split(rrs,",")
	for _, v := range rrsList {
		upfunc(v)
	}
}

func main() {
	// 初始化配置
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	iniConf, err := ini.Load(configPath)
	if err != nil {
		panic(err)
	}
	if s := iniConf.Section("aliyun");s != nil {
		regionID = s.Key("region_id").String()
		accessKey = s.Key("access_key_id").String()
		accessKeySecret = s.Key("access_key_secret").String()
	}
	if s := iniConf.Section("domain");s!=nil{
		rrs = s.Key("rr").String()
		domain = s.Key("name").String()
	}
	fmt.Println(regionID,accessKey,accessKeySecret,domain,rrs)
	// 创建ecsClient实例
	for {
		time.Sleep(time.Second * 10)
		localIp := get_external()
		if localIp == "" || netIp == localIp {
			continue
		}
		netIp = localIp
		fmt.Println(netIp)
		replace(netIp)
	}
}
