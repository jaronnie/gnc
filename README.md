# gnc
golang nc tool implement nc

## 下载

```shell
go install github.com/jaronnie/gnc@latest
```

## Usage

```shell
gnc -h
```

## 检查端口冲突

```shell
gnc -p 9102,3000,9100
```

如果存在端口冲突输出如下

```shell
$ gnc -lp 9102,3000,9100     
conflict ports: 3000,9100
```

## 检查端口冲突并且未冲突时监听端口

```shell
gnc -lp 3000,8000

gnc -lp 3000/tcp,8000/udp

curl localhost:3000
# output: success

nc -u 127.0.0.1 8000
```

## 设置存活时间(单位秒)

```shell
gnc -lp 3000,8000 -t 5

# 5s 后输出如下
get a signal quit
close network successfully
```