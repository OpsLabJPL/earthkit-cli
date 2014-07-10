package commands

import (
	"fmt"
	"github.com/opslabjpl/earthkit-cli/pool"
	"github.com/opslabjpl/earthkit-cli/workspace"
	"os"
	"strconv"
)

func PoolCreateCommand(args []string) {
	diskSize := 20
	if len(args) < 2 {
		fmt.Println("You need to specify machine type and machine count")
		fmt.Println("Usage:", os.Args[0], "pool-create machine_type machine_count [disk_size]")
		return
	}
	if len(args) == 3 {
		diskSize, _ = strconv.Atoi(args[2])
	}

	machineType := args[0]
	machineCount, err := strconv.Atoi(args[1])

	if err != nil {
		panic(err)
	}

	ws := workspace.GetWorkspace(".")
	myPool := pool.New(ws)
	myPool.CreateMachines(machineType, machineCount, diskSize)
}

func PoolListCommand(args []string) {
	ws := workspace.GetWorkspace(".")
	myPool := pool.New(ws)
	myPool.ListMachines()
}

func PoolKillCommand(args []string) {
	ws := workspace.GetWorkspace(".")
	myPool := pool.New(ws)
	myPool.KillMachines(args)
}

func PoolStopCommand(args []string) {
	ws := workspace.GetWorkspace(".")
	myPool := pool.New(ws)
	myPool.StopMachines(args)
}
