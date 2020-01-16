# gozzo 尜舟

## 编译三个命令行工具

Windows 下双击运行项目目录下的 WinBuild.bat

Linux 下在项目目录下运行 make 

## 用途1：端口转发代理
```bash
# 将 6380 转发到 6379 端口
./proxy 6380:127.0.0.1:6379
```

## 用途2：服务中继，类似的技术有 LVS 或 nginx upstream
```bash
# 在下一个端口启动后端服务，转发到目标端口
./relay -f servers.toml -rs -v
```

## 用途3：TCP Server
```bash
# 使用配置文件启动服务
./server settings.toml
```

# 服务器配置

```
cat > /etc/systemd/system/myserver.service <<EOD
[Unit]
Description=A TCP Server
After=syslog.target network.target

[Service]
Environment=
ExecStart=/opt/gozzo/real_server -p 9876 -logdir="/opt/gozzo/logs/"
WorkingDirectory=/opt/gozzo
#PIDFile=/var/run/myserver.pid
LimitNOFILE=819200
LimitNPROC=819200
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=myserver
Restart=always

[Install]
WantedBy=multi-user.target
EOD


cat > /etc/rsyslog.d/daemon.conf <<EOD
#*.*;daemon.none,auth,authpriv.none     /var/log/syslog
#daemon.*                               -/var/log/daemon.log
:app-name, isequal, "myserver"      -/opt/gozzo/logs/server.log
EOD


cat >> /etc/security/limits.conf <<EOD

#<domain>      <type>  <item>         <value>
*        soft    nofile        819200
*        hard    nofile        819200
root     soft    nofile        819200
root     hard    nofile        819200

#<domain>      <type>  <item>         <value>
*        soft    nproc         819200
*        hard    nproc         819200
root     soft    nproc         819200
root     hard    nproc         819200

#<domain>      <type>  <item>         <value>
*        soft    sigpending         409600
*        hard    sigpending         409600
root     soft    sigpending         409600
root     hard    sigpending         409600
EOD

ulimit -SHn 819200 && ulimit -SHu 819200 && ulimit -SHi 409600


cat >> /etc/sysctl.conf <<EOD

fs.file-max = 819200
vm.max_map_count = 819200
kernel.pid_max = 204800
kernel.sysrq = 1

net.core.netdev_max_backlog = 32000
net.core.rmem_max = 16777216
net.core.somaxconn = 8192
net.core.wmem_max = 16777216

net.ipv4.conf.all.arp_announce=2
net.ipv4.conf.all.rp_filter=0
net.ipv4.conf.all.send_redirects = 1
net.ipv4.conf.default.arp_announce = 2
net.ipv4.conf.default.rp_filter=0
net.ipv4.conf.default.send_redirects = 1
net.ipv4.conf.lo.arp_announce=2

net.ipv4.ip_forward = 1
net.ipv4.ip_local_port_range = 5001  65535
net.ipv4.icmp_echo_ignore_broadcasts = 1 # 避免放大攻击
net.ipv4.icmp_ignore_bogus_error_responses = 1 # 开启恶意icmp错误消息保护

net.ipv4.tcp_fin_timeout = 30
net.ipv4.tcp_keepalive_time = 1800
net.ipv4.tcp_max_syn_backlog = 1024
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.tcp_max_tw_buckets = 5000
net.ipv4.tcp_rmem = 4096 87380 16777216

net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_timestamps = 1
#net.ipv4.tcp_tw_recycle = 1
net.ipv4.tcp_tw_reuse = 1
#net.ipv4.tcp_tw_timeout = 3
net.ipv4.tcp_wmem = 4096 65536 16777216
EOD

sysctl -p
```