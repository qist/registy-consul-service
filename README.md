# registy-consul-service

##### 教程

```
配置文件地址参考： https://github.com/qist/registy-consul-service/blob/master/conf/consul.yaml 
linux 版本下载 
wget https://github.com/qist/registy-consul-service/releases/download/release/consulR
windows 版本下载
https://github.com/qist/registy-consul-service/releases/download/release/consulR.exe
windows 创建服务 使用nssm
C:\nssm-2.24\win64\nssm.exe install consulR C:\consulR\consulR.exe -confpath=C:\consulR\consul.yaml
启动服务
C:\nssm-2.24\win64\nssm.exe start consulR
$ ./consulR -confpath=<path-to-conf>
支持 tcp http 检查
linux 静态编译
CGO_ENABLED=0 go build  -o consulR -a -ldflags  '-extldflags "-static"' .
windows 版本编译
env GOOS=windows GOARCH=amd64 go build -o consulR.exe
mac 版本编译
env GOOS=darwin GOARCH=amd64 go build -o consulR-darwin
# 注册监控自动发现建议 CheckDeregisterCriticalServiceAfter 设置为 false 不然prometheus 不会报警因为consul 会删除注册信息
配置文件目录： conf
BlackboxExporter.yaml Blackbox-Exporter 注册发现配置 建议直接应用即可结束 consulR 进程
prometheus.yml prometheus 的服务发现配置 参考
consul.yaml 注册 本地 Exporter CheckDeregisterCriticalServiceAfter 设置为 false 可以不开启 consulR 进程
consul.yaml  作为业务注册时 建议 CheckDeregisterCriticalServiceAfter  设置为 true 这样业务故障consul 删除注册 业务恢复正常自动注册到consul 启动文件参考systemd/consulR@.service
心跳检测时间1分钟检测一次
```


