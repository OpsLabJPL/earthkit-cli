package cloudrun

import (
	// "bytes"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/opslabjpl/earthkit-cli/config"
	"github.com/opslabjpl/earthkit-cli/pool"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"github.com/opslabjpl/etcdq.git/etcdq"
	// "math/rand"
	"net/http"
	// "net/url"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

func New() *CloudRun {
	// Get list of nodes in the cluster
	ws := workspace.GetWorkspace(".")
	myPool := pool.New(ws)
	_, instances := myPool.MachinesReady()

	if len(instances) < 1 {
		panic("There is no running instance in your pool.")
	}

	// Create etcd client
	machines := make([]string, len(instances))
	for i, instance := range instances {
		machines[i] = "http://" + instance.IPAddress + ":4001"
	}
	etcd := etcd.NewClient(machines)

	cloudRun := &CloudRun{EtcdClient: etcd}
	return cloudRun
}

func Run(image string, cmd []string, filesetName string, patterns []string, workingDir string) {
	job := buildJobRequest(image, cmd, filesetName, patterns, workingDir)
	cloudRun := New()
	cloudRun.RunJob(job)
}

// TODO: support streaming
func getLog(worker etcdq.Worker, containerId string) {
	dockerEndpoint := "http://" + worker.PublicIp + ":4243/"

	resp, _ := http.Get(dockerEndpoint + "containers/" + containerId + "/logs?stdout=1&stderr=1")

	logs, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", logs)
}

// For some reason, using the go docker client to get the logs doesn't work
// for isce container's logs
func getLogOld(worker etcdq.Worker, containerId string) {
	dockerEndpoint := "http://" + worker.PublicIp + ":4243"
	dockerClient, err := docker.NewClient(dockerEndpoint)
	if err != nil {
		log.Fatal(err.Error())
	}

	println("Attaching to docker container", containerId)
	var buf bytes.Buffer

	err = dockerClient.AttachToContainer(docker.AttachToContainerOptions{
		Container:    containerId,
		OutputStream: &buf,
		Logs:         true,
		Stdout:       true,
		Stderr:       true,
	})
	if err != nil {
		panic(err.Error())
	}
	println("=====================================================")
	println("Job Output:")
	println(buf.String())
}

func buildJobRequest(image string, cmd []string, filesetName string, patterns []string, workingDir string) etcdq.Job {
	// Using private registry
	dockerRegistry := *config.DOCKER_REGISTRY
	if dockerRegistry != "" {
		image = dockerRegistry + "/" + image
	}

	timestamp, _ := json.Marshal(time.Now())
	tsStr := strings.Trim(string(timestamp), "\"")
	outFilesetName := filesetName + "_" + tsStr

	containerConf := docker.Config{
		Image:      image,
		Cmd:        cmd,
		WorkingDir: workingDir,
	}
	container := docker.Container{
		Config: &containerConf,
	}
	inFileSet := etcdq.Fileset{
		Name:     filesetName,
		Patterns: patterns,
	}
	outFileSet := etcdq.Fileset{
		Name:     outFilesetName,
		Patterns: patterns,
	}
	fileSet := etcdq.JobFileset{
		Input:  &inFileSet,
		Output: &outFileSet,
	}
	state := etcdq.State{
		State: "QUEUED",
	}
	job := etcdq.Job{
		Container: &container,
		Fileset:   &fileSet,
		States:    []etcdq.State{state},
	}
	return job
}

func printJobStatus(jobId string, job etcdq.Job) {
	println("job_id:", jobId, "| input:", job.Fileset.Input.Name, "| output:", job.Fileset.Output.Name, "| status:", job.States[len(job.States)-1].State, job.States[len(job.States)-1].Step)
}
