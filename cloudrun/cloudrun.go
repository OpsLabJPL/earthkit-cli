package cloudrun

import (
	"encoding/json"
	"fmt"
	"github.com/opslabjpl/etcdq.git/etcdq"
	// "net/http"
)

func (this *CloudRun) RunJob(job etcdq.Job) {
	etcd := this.EtcdClient
	jobJson, _ := json.Marshal(job)

	resp, _ := etcd.CreateInOrder("/jobs", string(jobJson), 0)
	println("Job ID:", resp.Node.Key)
	println("Result fileset: " + job.Fileset.Output.Name)
}

func (this *CloudRun) GetJobStatuses() {
	key := "/jobs"
	etcd := this.EtcdClient
	resp, _ := etcd.Get(key, true, false)

	if resp == nil || resp.Node == nil || resp.Node.Nodes == nil {
		println("There is no job history.")
		return
	}

	for _, node := range resp.Node.Nodes {
		job := etcdq.Job{}
		err := json.Unmarshal([]byte(node.Value), &job)

		if err != nil {
			fmt.Println("error:", err)
		}

		printJobStatus(node.Key, job)
	}
}

func (this *CloudRun) GetJobStatus(jobId string) {
	etcd := this.EtcdClient
	key := jobId
	resp, _ := etcd.Get(key, true, false)

	if resp == nil || resp.Node == nil {
		println("There is no job info for the given id")
		return
	}
	job := etcdq.Job{}
	err := json.Unmarshal([]byte(resp.Node.Value), &job)
	if err != nil {
		panic(err.Error())
	}
	printJobStatus(jobId, job)

	worker := this.getWorker(&job)
	getLog(worker, job.Container.ID)
}

func (this *CloudRun) getWorker(job *etcdq.Job) (worker etcdq.Worker) {
	etcd := this.EtcdClient
	resp, _ := etcd.Get("/workers/"+job.Owner, true, false)

	if resp == nil || resp.Node == nil {
		println("There is no worker info for the given job")
		return
	}

	err := json.Unmarshal([]byte(resp.Node.Value), &worker)
	if err != nil {
		panic(err.Error())
	}
	return
}
