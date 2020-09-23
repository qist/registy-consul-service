# registy-consul-service

##### 教程

```
$ ./consulR -confpath=<path-to-conf>
支持 tcp http 检查
linux 静态编译
CGO_ENABLED=0 go build  -o consulR -a -ldflags  '-extldflags "-static"' .
windows 版本编译
env GOOS=windows GOARCH=amd64 go build -o consulR.exe
```


