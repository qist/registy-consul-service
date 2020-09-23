# registy-consul-service

##### 教程

```
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
```


