# aliyun-ddns
a一个ddns小工具 用来给家里的动态IP提供ddns绑定
默认build了 arm liunx 版本
```
glide update
go build
```
修改 conf.ini
```
ddns -c conf.ini &
```
