
环境信息记录
外网访问：
IP:124.88.174.125
端口：50081
内网IP:10.24.250.81
账号：root
密码：S!pR4in#BlossOms=2025

10.24.21.149    XJ01    117.145.184.180:20001  --ok.  0904 软links错误
10.24.21.151    XJ02    124.88.174.125:18989
10.24.21.152    XJ03    222.81.173.56:30002   --ok   0905
10.24.21.153    XJ04    116.178.134.133:30003 --ok. 0903
10.24.21.155    XJ05    222.81.173.56:30004



1. 要求，按照项目、语言、站点重命名文件
  1. 命名格式：供应商名称首字母大写-日期年月日-数据类别-语言类型-数据站源
  2. 例子YZ-20250811-ASMR-Chinese-YTB
  3. 现有的项目：ASMR、DJ、Storyline、GKLT、Opentalk
  4. 最终云上呈现：ASMR/YZ-20250811-ASMR-Chinese-YTB/id/
    1. id.mp4
    2. id.meta
  5. 语言：中文、英文

#!/bin/bash


innerhost=xj04
innerport=30003

apt-get update
apt-get install qbittorrent-nox  -y

mkdir -p /data/bt-download/$innerhost/data
ln -s  /data/bt-download/$innerhost/data  /root/Downloads

service_file=/etc/systemd/system/bt-download.service

cat <<EOF > "$service_file"

[Unit]
Description=qBittorrent-nox service
After=network.target

[Service]
# 设置环境变量
Environment=HOST=$innerhost

# 启动用户（根据实际情况修改，建议使用非root用户）
User=root
Group=root

# 启动命令
ExecStart=/usr/bin/qbittorrent-nox --webui-port=$innerport --save-path="/data/bt-download/$innerhost/data"

# 进程退出后自动重启
Restart=always
RestartSec=5

# 限制资源使用（可选）
LimitNOFILE=65536

[Install]
# 设置为多用户模式下的自启动
WantedBy=multi-user.target

EOF

systemctl enable bt-download.service
systemctl restart bt-download.service
systemctl status bt-download.service


 $ ssh-keygen -t rsa -b 4096 -C "239566803@qq.com"
Generating public/private rsa key pair.
Enter file in which to save the key (/Users/winnie/.ssh/id_rsa):
Enter passphrase (empty for no passphrase):
Enter same passphrase again:
Your identification has been saved in /Users/winnie/.ssh/id_rsa
Your public key has been saved in /Users/winnie/.ssh/id_rsa.pub
The key fingerprint is:
SHA256:r7qNv/SgV+gSvX1me6RHR5nsaXSbz+F9k3qcgqyob3I 239566803@qq.com
The key's randomart image is:
+---[RSA 4096]----+
|                 |
|                 |
|              . o|
|               =o|
|       .S.    o.=|
|      . o..   oB.|
|       oo+o .++o*|
|     ..E=+.o=.oB=|
|     .@O=oo+.=+ o|
+----[SHA256]-----+


bq优化
vim /root/.config/qBittorrent/qBittorrent.conf

[Preferences]
...
# 预分配模式 (0: 稀疏文件, 1: 完全分配, 2: 禁用)
Session\PreallocationMode=0
...
# 另一个相关设置：为快速resume数据预分配（默认true即可）
Session\PreallocateAllFiles=true