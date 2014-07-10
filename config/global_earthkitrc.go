package config

import (
	"flag"
	"github.com/rakyll/globalconf"
	"github.com/opslabjpl/goamz/aws"
	"os"
	"path"
	"time"
)

var DOCKER_PATH = flag.String("docker_path", "/usr/bin/docker", "Path to the docker binary (usually /usr/bin/docker).")
var DOCKER_HTTP_PORT = flag.Int("docker_http_port", -1, "HTTP port that the docker daemon is listening on (defaults to UNIX socket if unspecified).")
var AMI = flag.String("ami", "todo", "AMI to use")
var SECURITY_GROUP = flag.String("security_group", "todo", "Security group to use")
var SUBNET = flag.String("subnet", "todo", "Subnet to use")
var EC2_KEY_PAIR = flag.String("ec2_key_pair", "todo", "EC2 keyname to use")
var AWS_ACCESS_KEY = flag.String("aws_access_key", "todo", "AWS ACCESS KEY to use")
var AWS_SECRET_KEY = flag.String("aws_secret_key", "todo", "AWS SECRET KEY to use")
var S3_BUCKET = flag.String("s3_bucket", "earthkit-cli", "S3 bucket to use")
var S3_KEY_PREFIX = flag.String("s3_key_prefix", ".earthkit", "S3 key prefix to use")
var AWS_REGION = flag.String("aws_region", "us-gov-west-1", "AWS Region to use")
var DOCKER_REGISTRY = flag.String("docker_registry", "", "Docker registry to use")
var ETCDQ_S3_PATH = flag.String("etcdq_s3_path", "bin/ubuntu/etcdq", "S3 path to etcdq binary")
var DOCKER_CONF_S3_PATH = flag.String("docker_conf_s3_path", "conf/docker.conf", "S3 path to docker conf file")
var DOCKER_INSTALL_S3_PATH = flag.String("docker_install_s3_path", "bin/get.docker.io.sh", "S3 path to script for isntall docker")
var DATA_DIR = flag.String("data_dir", "/mnt/data/earthkit", "Directory to mount EBS volume to for cloud processing")
var CACHE_LIMIT = flag.Int64("cache_limit", 5368709120, "Cache limit (in bytes)")
var EKIT_IMG = flag.String("earthkit_img", "earthkit-cli", "Docker image containing earhtkit-cli command")
var Verbose = flag.Bool("v", false, "enables verbose output")
var Region aws.Region = aws.Regions[*AWS_REGION]

func WorkspacePrefix(workspace string) string {
	return path.Join(*S3_KEY_PREFIX, workspace)
	// return path.Join(S3KeyPrefix, workspace)
}

func AWSAuth() (auth aws.Auth) {
	auth, err := aws.GetAuth(*AWS_ACCESS_KEY, *AWS_SECRET_KEY, "", time.Time{})

	if err != nil {
		auth, err = aws.EnvAuth()
		if err != nil {
			panic(err.Error())
		}
	}
	return
}

func Load() {
	homeDir := os.Getenv("HOME")
	opts := globalconf.Options{Filename: homeDir + "/.earthkitrc"}
	conf, err := globalconf.NewWithOptions(&opts)
	if err != nil {
		// Just parse command line arguments at the root level if the configuration file doesn't exist
		flag.Parse()
		return
	}
	conf.ParseAll()
	Region = aws.Regions[*AWS_REGION]
}
