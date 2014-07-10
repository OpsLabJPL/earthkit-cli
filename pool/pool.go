package pool

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/goamz.git/ec2"
	"github.com/opslabjpl/goamz.git/s3"
	"github.com/opslabjpl/goprovision.git"
	"strings"
	"time"
)

func (pool *Pool) CreateMachines(machineType string, machineCount int, diskSize int) (instances []ec2.Instance) {

	// If there is no machines available, generate new discovery token
	existingInstances := pool.GetMachines()
	genDisToken := true
	for _, instance := range existingInstances {
		// Skip machines that are stopping/stopped or terminated
		if instance.State.Code > 16 {
			continue
		}
		genDisToken = false
		break
	}
	if genDisToken {
		pool.Workspace.SetUpDiscoveryUrl(false)
	}

	blockDeviceMappings := ec2.BlockDeviceMapping{
		DeviceName: "/dev/sdf1",
		VolumeSize: int64(diskSize),
	}

	// standard ec2 options as required by goamz ec2
	ec2Opts := ec2.RunInstancesOptions{
		ImageId:                  *config.AMI,
		InstanceType:             machineType,
		MinCount:                 machineCount,
		KeyName:                  *config.EC2_KEY_PAIR,
		SubnetId:                 *config.SUBNET,
		SecurityGroups:           []ec2.SecurityGroup{{Id: *config.SECURITY_GROUP}},
		UserData:                 []byte(pool.genUserData()),
		AssociatePublicIpAddress: true,
		BlockDeviceMappings:      []ec2.BlockDeviceMapping{blockDeviceMappings},
	}

	if pool.Workspace.Name == "" {
		panic("Unable to find workspace name. Make sure you run the program from within a workspace.")
	} else {
		println("Creating pool machines for", pool.Workspace.Name)
	}

	tags := []ec2.Tag{{"workspace", pool.Workspace.Name}, {"Name", pool.Workspace.Name}}

	// additional options
	provOpts := goprovision.ProvOpts{
		Tags:               tags,
		TagAttachedVolumes: false,
	}

	instances, err := pool.Provisioner.CreateInstances(ec2Opts, provOpts)
	if err != nil {
		panic(err.Error())
	}
	return instances
}

func (pool *Pool) StopMachines(targets []string) {
	instances, readyInstances := pool.MachinesReady()
	instancesToStop := filterTargetMachines(targets, instances)

	newCount := len(readyInstances) - len(instancesToStop)
	if newCount < 3 && newCount > 0 {
		fmt.Println("Earthkit cannot handle your request because it will result in a cluster of less than 3 machines. If you want to create a cluster of less than 3 machines, you must first delete the current one.")
		return
	}

	fmt.Println("Stopping", instancesToStop)
	for _, instanceId := range instancesToStop {
		pool.Provisioner.EC2.StopInstances(instanceId)
	}

	// TODO: remove machine from etcd discovery peer list
}

func (pool *Pool) KillMachines(targets []string) {
	instances, readyInstances := pool.MachinesReady()
	instancesToKill := filterTargetMachines(targets, instances)

	newCount := len(readyInstances) - len(instancesToKill)
	if newCount < 3 && newCount > 0 {
		fmt.Println("Earthkit cannot handle your request because it will result in a cluster of less than 3 machines. If you want to create a cluster of less than 3 machines, you must first delete the current one.")
		return
	}

	fmt.Println("Terminating", instancesToKill)
	pool.Provisioner.EC2.TerminateInstances(instancesToKill)

	// TODO: remove machine from etcd discovery peer list
}

// Printing out list of machines that are in the pool.
// Does not print out machines that have status 'terminated'
func (pool *Pool) ListMachines() {
	instances := pool.GetMachines()
	for _, instance := range instances {
		status := getStatus(instance)

		// Don't display terminated machines
		if status == "terminated" {
			continue
		}

		fmt.Println(instance.InstanceId, "|", status, "|", instance.PrivateIPAddress, "|", instance.IPAddress)
	}
}

// Return a pair of slices. First slice is a list of all instances
// second slice is a list of instances that are up and running with etcd and docker
func (pool *Pool) MachinesReady() (instances []ec2.Instance, readyInstances []ec2.Instance) {
	instances = pool.GetMachines()

	// see which ones are up and running with docker and etcd
	for _, instance := range instances {
		if etcdAndDockerUp(instance.IPAddress) {
			readyInstances = append(readyInstances, instance)
		}
	}
	return
}

// Get list of machines that are in the pool. The machines are not
// necessary fully running (e.g. being created)
func (pool *Pool) GetMachines() (instances []ec2.Instance) {
	filter := ec2.NewFilter()
	filter.Add("tag-value", pool.Workspace.Name)
	filter.Add("tag-key", "workspace")
	instances, _ = pool.Provisioner.Instances(nil, filter)
	return
}

func filterTargetMachines(targets []string, instances []ec2.Instance) (instancesIds []string) {
	instancesIdsMap := make(map[string]bool)
	for _, instance := range instances {
		instancesIdsMap[instance.InstanceId] = false
	}

	all := false
	for _, target := range targets {
		if target == "all" {
			all = true
		} else if _, exists := instancesIdsMap[target]; exists {
			instancesIdsMap[target] = true
		} else {
			println("Skipping " + target + "because it is not in your workspace.")
		}
	}

	for instance, todelete := range instancesIdsMap {
		if all == true || todelete == true {
			instancesIds = append(instancesIds, instance)
		}
	}
	return
}

func (pool *Pool) genUserData() (userData string) {
	userData = UserData

	auth := config.AWSAuth()
	myS3 := s3.New(auth, config.Region)
	bucket := myS3.Bucket(*config.S3_BUCKET)

	discoveryURL := string(pool.Workspace.Remote().GetDiscoveryURL())

	dockerURL := bucket.SignedURL(*config.DOCKER_INSTALL_S3_PATH, time.Now().Add(60*time.Minute))
	etcdqURL := bucket.SignedURL(*config.ETCDQ_S3_PATH, time.Now().Add(60*time.Minute))
	dockerConfURL := bucket.SignedURL(*config.DOCKER_CONF_S3_PATH, time.Now().Add(60*time.Minute))

	workspaceName := pool.Workspace.Name
	earthkitImg := *config.DOCKER_REGISTRY + "/" + *config.EKIT_IMG

	userData = strings.Replace(userData, "__DOCKER_URL__", dockerURL, -1)
	userData = strings.Replace(userData, "__DOCKER_CONF_URL__", dockerConfURL, -1)
	userData = strings.Replace(userData, "__EARTHKIT_IMG__", earthkitImg, -1)

	userData = strings.Replace(userData, "__ETCDQ_URL__", etcdqURL, -1)
	userData = strings.Replace(userData, "__DISCOVERY_URL__", discoveryURL, -1)

	userData = strings.Replace(userData, "__AWS_ACCESS_KEY__", *config.AWS_ACCESS_KEY, -1)
	userData = strings.Replace(userData, "__AWS_SECRET_KEY__", *config.AWS_SECRET_KEY, -1)
	userData = strings.Replace(userData, "__S3_BUCKET__", *config.S3_BUCKET, -1)
	userData = strings.Replace(userData, "__AWS_REGION__", *config.AWS_REGION, -1)

	userData = strings.Replace(userData, "__WORK_SPACE__", workspaceName, -1)
	userData = strings.Replace(userData, "__DATA_DIR__", *config.DATA_DIR, -1)
	return
}
