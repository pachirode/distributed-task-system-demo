# 分布式任务系统

前端页面提交任务到数据库的任务表中，系统定时扫描任务表，取出待执行的任务，根据任务配置到 `K8S` 中拉起 `Job` 资源对象去真正的执行任务
取出已经开始执行的任务，获取当前任务的实时状态，并会写到数据库中
任务标记为失败或者成功，则任务最终结束

### 使用
1. 创建一个 `k8s` 集群
```bash
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.19.0/kind-linux-amd64
chmod +x ./kind

kind create cluster -n test-k8s 
```
2. 拉取镜像
```bash
docker pull busybox
docker pull alpine

kind load docker-image busybox:latest
kind load docker-image alpine:latest

kubectl create namespace demo
```

3. 运行项目
```bash
go run cmd/system-watch/main.go
```

