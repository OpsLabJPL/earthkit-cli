package pool

const UserData = `#!/bin/bash

# install docker
curl "__DOCKER_URL__" | bash

# fetch docker conf
wget "__DOCKER_CONF_URL__" -O /etc/init/docker.conf

# rerun docker with new config 
service docker restart

mkdir -p /cescre/logs/
mkdir /cescre/bin/
mkdir /etc/ekit/
mkdir /etc/etcd/
mkdir /cescre/etcd_data

# Download and install etcd
wget https://github.com/coreos/etcd/releases/download/v0.3.0/etcd-v0.3.0-linux-amd64.tar.gz
tar xzvf etcd-v0.3.0-linux-amd64.tar.gz -C /cescre/
rm etcd-v0.3.0-linux-amd64.tar.gz

# setup ekit conf
cat << EOF  > /etc/ekit/earthkitrc
aws_access_key = __AWS_ACCESS_KEY__
aws_secret_key = __AWS_SECRET_KEY__
s3bucket = __S3_BUCKET__
s3keyprefix = .earthkit
EOF

# setup etcd conf
pub_addr=` + "`" + `curl -s http://169.254.169.254/latest/meta-data/public-ipv4` + "`" + `
addr=` + "`" + `curl http://169.254.169.254/latest/meta-data/local-ipv4` + "`" + `
discovery=__DISCOVERY_URL__
name=` + "`" + `curl http://169.254.169.254/latest/meta-data/local-hostname` + "`" + `

cat << EOF  > /etc/etcd/etcd.conf
discovery = "$discovery"
addr = "$pub_addr:4001"
data_dir = "/cescre/etcd_data"
name = "$name"

[peer]
addr = "$addr:7001"
EOF

# start etcd, using conf defined in /etc/etcd/etcd.conf
/cescre/etcd-v0.3.0-linux-amd64/etcd >> /cescre/logs/etcd.log  2>&1 &

# Download etcq daemon worker
wget "__ETCDQ_URL__" -O /cescre/bin/etcdq
chmod a+x /cescre/bin/etcdq

# mount ebs volume
# TODO: Remove hardcode of device name and mount dir
sudo mkfs -t ext4 /dev/xvdf1
sudo mkdir /mnt/data
sudo mount /dev/xvdf1  /mnt/data
sudo mkdir /mnt/data/earthkit
sudo chmod 1777 /mnt/data/earthkit
sleep 5

DATA_DIR=__DATA_DIR__ EARTHKIT_IMG=__EARTHKIT_IMG__ AWS_ACCESS_KEY=__AWS_ACCESS_KEY__ AWS_SECRET_KEY=__AWS_SECRET_KEY__ AWS_REGION=__AWS_REGION__ S3_BUCKET=__S3_BUCKET__ WORKSPACE=__WORK_SPACE__ /cescre/bin/etcdq >> /cescre/logs/etcdq.log  2>&1 &

`
