package consul

import (
	"encoding/json"
	conf "exec/config"
    "crypto/tls"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ADI interface {
	GetAddrs() string
	GetAllAddr() string
	GetConsulAddr() string
	CheckAddrs(tag string, addr string) (bool, error)
	ConsulRegister()
}

type Addresses struct{}

type CatalogService struct {
	//获取的CatalogService的数据
	Address        string `json: "Address"`
	Datacenter     string `json: "Datacenter"`
	ServiceID      string `json: "ServiceID"`
	ServiceName    string `json: "ServiceName"`
	ServiceAddress string `json: "ServiceAddress"`
}

type CatalogServices []CatalogService

func GetAddrs() string {
	//获取出口网卡IP
	if conf.GetConf().Service.Address == "" {
		conn, err := net.Dial("udp", conf.GetConf().System.FindAddress)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
			panic("error" + err.Error())
		}
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String()
	}
	return conf.GetConf().Service.Address
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("you are visiting health check api"))
}

func IsToken() bool {
	v := conf.GetConf().Consul.Token
	if v != "" {
		return true
	} else {
		return false
	}
}

func SearchIssues(HttpAddress string, ApiAddress string) ([]string, []string, error) {
	//Get方式获取Json反序列化成CatalogService结构体
	//读取ServiceID和其所在的节点IP
	logrus.WithFields(logrus.Fields{}).Info("Geting Data From " + "http://" + HttpAddress + ApiAddress + "...")
	resp, err := http.Get("http://" + HttpAddress + ApiAddress + "?token=" + conf.GetConf().Consul.Token)
	if err != nil {
		logrus.Error(err.Error())
		return nil, nil, err
	}
	logrus.WithFields(logrus.Fields{}).Info("Geting Data From " + "http://" + HttpAddress + ApiAddress + " Successed...")
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	var result CatalogServices
	err = json.Unmarshal(content, &result)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
		return nil, nil, err
	}
	SvcIDs := make([]string, 0)
	SvcADs := make([]string, 0)
	for i := 0; i < len(result); i++ {
		SvcADs = append(SvcADs, []string{result[i].Address + ":8500"}...)
		SvcIDs = append(SvcIDs, []string{result[i].ServiceID}...)
	}
	return SvcIDs, SvcADs, err
}

func (GCA *Addresses) GetConsulAddr() string {
	//获取随机种子取得随机IP
	randaddr := rand.New(rand.NewSource(time.Now().UnixNano()))
	consuladdress := strings.Split(conf.GetConf().Consul.Address, ",")
	consulregisty := consuladdress[randaddr.Intn(3)]
	return consulregisty
}

func (GCA *Addresses) GetAllAddr() []string {
	//获取所有Consul地址
	consuladdress := strings.Split(conf.GetConf().Consul.Address, ",")
	return consuladdress
}

func (CA *Addresses) CheckAddrs(tag string, addr string) error {
	// 删除注册所有信息
	var conf *conf.Config
	config2, err := conf.GetConfig()
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
	}
	logrus.WithFields(logrus.Fields{}).Info("删除所有注册的信息...")
	for i := 0; i < len(CA.GetAllAddr()); i++ {
		config := consulapi.DefaultConfig()
		config.Address = CA.GetAllAddr()[i]
		config.Token = config2.Consul.Token
		var client *consulapi.Client
		client, err = consulapi.NewClient(config)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
			panic("consul client error" + err.Error())
		}
		err = client.Agent().ServiceDeregister(tag + "-" + addr)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
			return err
		}
	}
	return nil
}

func (CR *Addresses) ConsulRegister(addr string) {
	// 创建连接consul服务配置
	config := consulapi.DefaultConfig()
	config.Address = addr
	config.Token = conf.GetConf().Consul.Token
	client, err := consulapi.NewClient(config)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
		panic("consul client error" + err.Error())
	}
	logrus.WithFields(logrus.Fields{}).Info("建立了新的consul连接...")
	// 创建注册到consul的服务到
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = conf.GetConf().Service.Tag + "-" + GetAddrs() + "-" + conf.GetConf().Service.Port
	registration.Name = conf.GetConf().Service.Tag
	var port int
	port, err = strconv.Atoi(conf.GetConf().Service.Port)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
	}
	registration.Port = port
	registration.Tags = []string{conf.GetConf().Service.Tag}
    if conf.GetConf().Service.Address == "" || conf.GetConf().Consul.CheckType == "tcp" {
	registration.Address = GetAddrs()
    } else if conf.GetConf().Service.Address != "" && conf.GetConf().Service.Port == "80" || conf.GetConf().Service.Port == "443" {
    registration.Address = conf.GetConf().Consul.CheckType  + "://" + GetAddrs() + conf.GetConf().Consul.CheckHealth
    }  else {
    registration.Address = conf.GetConf().Consul.CheckType  + "://" + GetAddrs() + ":" +  strconv.Itoa(port) + conf.GetConf().Consul.CheckHealth
    }
	// 增加consul健康检查回调函数
	check := new(consulapi.AgentServiceCheck)
    if conf.GetConf().Consul.CheckType == "http" {
	    check.HTTP = fmt.Sprintf("http://%s:%d%s?token=%v", GetAddrs(), registration.Port, conf.GetConf().Consul.CheckHealth, conf.GetConf().Consul.Token)
        check.Timeout = conf.GetConf().Consul.CheckTimeout
	    check.Interval = conf.GetConf().Consul.CheckInterval
    }else if conf.GetConf().Consul.CheckType == "https" {
	    check.HTTP = fmt.Sprintf("https://%s:%d%s?token=%v", GetAddrs(), registration.Port, conf.GetConf().Consul.CheckHealth, conf.GetConf().Consul.Token)
        check.Timeout = conf.GetConf().Consul.CheckTimeout
        check.TLSSkipVerify = true 
	    check.Interval = conf.GetConf().Consul.CheckInterval
    }else if conf.GetConf().Consul.CheckType == "tcp" {
        check.TCP = fmt.Sprintf("%s:%d", GetAddrs(), registration.Port)
        check.Timeout = conf.GetConf().Consul.CheckTimeout
	    check.Interval = conf.GetConf().Consul.CheckInterval
     }
	if conf.GetConf().Consul.CheckDeregisterCriticalServiceAfter == false {
		logrus.WithFields(logrus.Fields{}).Info("未开启自动删除注册服务...")
	} else if conf.GetConf().Consul.CheckDeregisterCriticalServiceAfterTime != "" {
		check.DeregisterCriticalServiceAfter = conf.GetConf().Consul.CheckDeregisterCriticalServiceAfterTime
	} else {
		logrus.WithFields(logrus.Fields{}).Fatal("未输入自动删除时间...")
	}
	//check.DeregisterCriticalServiceAfter = "10s" // 故障检查失败30s后 consul自动将注册服务删除
	registration.Check = check
	logrus.WithFields(logrus.Fields{}).Info("New Consul-Service Health Heartbeat Created...")
	// 注册服务到consul
	logrus.WithFields(logrus.Fields{}).Info("Service Registying...")
	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
		panic("error" + err.Error())
	}
}

func (CS *Addresses) CheckSorted(ServiceName string) (string, error) {
	//计算出注册数为0的节点
	_, SvcADs, err := SearchIssues(CS.GetConsulAddr(), "/v1/catalog/service/"+ServiceName)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
		return "", err
	}
	logrus.WithFields(logrus.Fields{}).Info("Computing Registied node...")
	dict := make(map[string]int)
	for _, num := range SvcADs {
		dict[num] = dict[num] + 1
	}
	for i, v := range CS.GetAllAddr() {
		if dict[v] == 0 {
			return CS.GetAllAddr()[i], nil
		}
	}
	//计算出最少注册数的节点
	mixIP := ""
	for k, mix := range dict {
		mixIP = k
		for k1, v1 := range dict {
			if v1 < mix {
				mixIP = k1
				mix = v1
			}
		}
		break
	}
	return mixIP, nil
}

func GetSvcCode() bool {
    if conf.GetConf().Consul.CheckType == "http" {
	   u, _ := url.Parse("http://" + GetAddrs() + ":" + conf.GetConf().Service.Port + conf.GetConf().Consul.CheckHealth)
	   q := u.Query()
	   u.RawQuery = q.Encode()
	   res, err := http.Get(u.String())
	   if err != nil {
	   	fmt.Println("0")
	   	return false
	   }
	   resCode := res.StatusCode
	   res.Body.Close()
	   if err != nil {
	   	fmt.Println("0")
	   	return false
	   }
	   if resCode == 200 {
	   	return true
	   }
   } else if conf.GetConf().Consul.CheckType == "https" {
	   u, _ := url.Parse("https://" + GetAddrs() + ":" + conf.GetConf().Service.Port + conf.GetConf().Consul.CheckHealth)
	   q := u.Query()
	   u.RawQuery = q.Encode()
       c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}}
	   res, err := c.Get(u.String())
	   if err != nil {
	   	fmt.Println("0")
	   	return false
	   }
	   resCode := res.StatusCode
	   res.Body.Close()
	   if err != nil {
	   	fmt.Println("0")
	   	return false
	   }
	   if resCode == 200 || resCode == 401 {
	   	return true
	   }
   } else if conf.GetConf().Consul.CheckType == "tcp" {
     timeout := time.Second
     conn, err := net.DialTimeout("tcp", net.JoinHostPort(GetAddrs(), conf.GetConf().Service.Port), timeout)
     if err != nil {
     fmt.Println("0")
     return false
     }
     if conn != nil {
     defer conn.Close()
     return true
     }
   }
  return false
}

func (CA *Addresses) CheckAddr(ServiceName string) error {
	//获取并过滤注册信息的ServiceID
	SvcIDs, _, err := SearchIssues(CA.GetConsulAddr(), "/v1/catalog/service/"+ServiceName)
	if err != nil {
		logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
		return err
	}
	var i int
	for _, v := range SvcIDs {
		if v == conf.GetConf().Service.Tag+"-"+GetAddrs() + "-" + conf.GetConf().Service.Port {
			i++
		}
	}
	logrus.WithFields(logrus.Fields{}).Info("Geted Registied ServiceID Addr...")
	//通过查询到的注册情况判定如何注册
	var addr string
	if i > 1 {
		CA.CheckAddrs(conf.GetConf().Service.Tag, GetAddrs())
		addr, err = CA.CheckSorted(conf.GetConf().Service.Tag)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
			return err
		}
		CA.ConsulRegister(addr)
		logrus.WithFields(logrus.Fields{}).Info("More than 1 ServiceID detected,Service Registied Success...")
	} else if i == 0 {
		addr, err = CA.CheckSorted(conf.GetConf().Service.Tag)
		if err != nil {
			logrus.WithFields(logrus.Fields{}).Fatal(err.Error())
			return err
		}
		CA.ConsulRegister(addr)
		logrus.WithFields(logrus.Fields{}).Info("None ServiceID detected,Service Registied Success...")
	} else {
		logrus.WithFields(logrus.Fields{}).Info("A Same ServiceID Detected...")
	}
	return nil
}
