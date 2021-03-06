package totoro

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"time"
	"totoro/util"
)

func initCtxAndCli() (context.Context, *client.Client) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli.NegotiateAPIVersion(ctx)

	return ctx, cli
}

func GetContainerIdByName(name string) string {
	_, cli := initCtxAndCli()

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		for _, nm := range container.Names {
			if nm == name {
				return container.ID
			}
		}
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
	return ""
}

func InspectContainerById(containerId string) (string, int) {
	ctx, cli := initCtxAndCli()

	info, err := cli.ContainerInspect(ctx, containerId)
	if err != nil || info.ContainerJSONBase == nil{
		util.PrintErr("[error] inspect container err")
		e := cli.Close()
		if e != nil {
			util.PrintInfo("[error] cli connection close err")
		}
		return "", 0
	}
	status := info.ContainerJSONBase.State

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
	return status.Status, status.ExitCode
}

func GetAppResourceInfo(containerId string) (float64, float64){
	ctx, cli := initCtxAndCli()

	info, er := cli.ContainerStats(ctx, containerId, false)
	if er != nil {
		util.PrintErr("[error] getAppInfo err")
		return 0.0, 0.0
	}
	if info.Body != nil {
		buf := new(bytes.Buffer)
		buf.ReadFrom(info.Body)
		var containerStats types.Stats
		err := json.Unmarshal(buf.Bytes(), &containerStats)
		if err != nil {
			util.PrintErr("[error] json unmarshal error")
			e := cli.Close()
			if e != nil {
				util.PrintInfo("[error] cli connection close err")
			}
			return 0.0, 0.0
		}

		/*
		* used_memory = `memory_stats.usage - memory_stats.stats.cache`
		* available_memory = `memory_stats.limit`
		* Memory usage % = `(used_memory / available_memory) * 100.0`
		* cpu_delta = `cpu_stats.cpu_usage.total_usage - precpu_stats.cpu_usage.total_usage`
		* system_cpu_delta = `cpu_stats.system_cpu_usage - precpu_stats.system_cpu_usage`
		* number_cpus = `lenght(cpu_stats.cpu_usage.percpu_usage)` or `cpu_stats.online_cpus`
		* CPU usage % = `(cpu_delta / system_cpu_delta) * number_cpus * 100.0`
		 */
		usedMemory := containerStats.MemoryStats.Usage - containerStats.MemoryStats.Stats["cache"]
		availableMemory := containerStats.MemoryStats.Limit
		memoryUsage := (float64(usedMemory) / float64(availableMemory)) * 100.0
		cpuDelta := containerStats.CPUStats.CPUUsage.TotalUsage - containerStats.PreCPUStats.CPUUsage.TotalUsage
		systemCpuDelta := containerStats.CPUStats.SystemUsage - containerStats.PreCPUStats.SystemUsage
		numberCpus := uint64(len(containerStats.CPUStats.CPUUsage.PercpuUsage))
		cpuUsage := (float64(cpuDelta) / float64(systemCpuDelta)) * float64(numberCpus) * 100.0
		//util.PrintInfo("[info] container:(%s) cpu_usage:%f memory_usage:%f", containerId, cpuUsage, memoryUsage)
		//util.PrintInfo("[info] cpuDelta: %d", cpuDelta)
		//util.PrintInfo("[info] systemCpuDelta: %d", systemCpuDelta)
		//util.PrintInfo("[info] numberCpus: %d", numberCpus)
		//util.PrintInfo("[info] usedMemory: %d", usedMemory)
		//util.PrintInfo("[info] availableMemory: %d", availableMemory)

		er := info.Body.Close()
		if er != nil {
			util.PrintInfo("[error] info body connection close err")
		}
		e := cli.Close()
		if e != nil {
			util.PrintInfo("[error] cli connection close err")
		}

		return cpuUsage, memoryUsage
	} else {
		er := info.Body.Close()
		if er != nil {
			util.PrintInfo("[error] info body connection close err")
		}
		e := cli.Close()
		if e != nil {
			util.PrintInfo("[error] cli connection close err")
		}
		util.PrintErr("[error] (cli.ContainerStats) return nil Body")
	}
	return 0.0, 0.0
}

func CreateContainerByImageName(containerName string, config *container.Config, hostConfig *container.HostConfig) string {
	ctx, cli := initCtxAndCli()

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	//util.PrintInfo("[info] container (%s) is running", containerName)

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
	return resp.ID
}

func UpdateContainerCpuShareById(containerId string, cpuShare int64) {
	ctx, cli := initCtxAndCli()

	_, err := cli.ContainerUpdate(ctx, containerId, container.UpdateConfig{
		Resources: container.Resources{
			CPUShares: cpuShare,
		},
	})
	if err != nil {
		util.PrintErr("[error] container (%s) resource update error", containerId)
		fmt.Println(err)
	} else {
		util.PrintInfo("[info] container updated")
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func UpdateContainerCpuQuotaById(containerId string, cpuQuota int64) {
	ctx, cli := initCtxAndCli()

	_, err := cli.ContainerUpdate(ctx, containerId, container.UpdateConfig{
		Resources: container.Resources{
			CPUPeriod: 100000,
			CPUQuota: cpuQuota,
		},
	})
	if err != nil {
		util.PrintErr("[error] container (%s) resource update error", containerId)
		fmt.Println(err)
	} else {
		util.PrintInfo("[info] container updated: cpuQuota %d", cpuQuota)
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func UpdateContainerCpuSetsById(containerId string, cpuSets string) {
	ctx, cli := initCtxAndCli()

	_, err := cli.ContainerUpdate(ctx, containerId, container.UpdateConfig{
		Resources: container.Resources{
			CpusetCpus: cpuSets,
		},
	})
	if err != nil {
		util.PrintErr("[error] container (%s) resource update error", containerId)
		fmt.Println(err)
	} else {
		//util.PrintInfo("[info] container cpuSets updated: cpuSets { %s }", cpuSets)
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func StopContainerById(containerId string, timeout time.Duration) {
	ctx, cli := initCtxAndCli()

	err := cli.ContainerStop(ctx, containerId, &timeout)
	if err != nil {
		util.PrintErr("[error] container (%s) stop error", containerId)
	} else {
		//util.PrintInfo("[info] container stopped")
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func KillContainerById(containerId string) {
	ctx, cli := initCtxAndCli()

	err := cli.ContainerKill(ctx, containerId, "")
	if err != nil {
		util.PrintErr("[error] container (%s) kill error", containerId)
	} else {
		//util.PrintInfo("[info] container killed")
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func StartContainerById(containerId string) {
	ctx, cli := initCtxAndCli()

	if err := cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{}); err != nil {
		panic(err)
	} else {
		//util.PrintInfo("[info] container started")
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}

func RemoveContainerById(containerId string) {
	ctx, cli := initCtxAndCli()

	err := cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{})
	if err != nil {
		util.PrintErr("[error] container (%s) remove error", containerId)
	} else {
		util.PrintInfo("[info] container removed")
	}

	e := cli.Close()
	if e != nil {
		util.PrintInfo("[error] cli connection close err")
	}
}