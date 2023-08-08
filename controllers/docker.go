package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/voidchef/devops/utils"
)

func GetContainers(c *gin.Context) {
	containers, err := utils.ListContainers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type Port struct {
		IP          string `json:"ip"`
		PrivatePort uint16 `json:"privatePort"`
		PublicPort  uint16 `json:"publicPort"`
		Type        string `json:"type"`
	}

	type Container struct {
		ID              string    `json:"id"`
		Name            string    `json:"name"`
		Image           string    `json:"image"`
		ImageID         string    `json:"imageID"`
		Command         string    `json:"command"`
		Created         time.Time `json:"created"`
		State           string    `json:"state"`
		Status          string    `json:"status"`
		Port            []Port    `json:"port"`
		SizeRw          int64     `json:"sizeRw"`
		SizeRootFs      int64     `json:"sizeRootFs"`
		NetworkMode     string    `json:"networkMode"`
		NetworkSettings []string  `json:"networkSettings"`
		Mounts          string    `json:"mounts"`
		Labels          []byte    `json:"labels"`
	}

	var containerList []Container

	for _, container := range containers {
		data := Container{}
		data.ID = container.ID
		data.Name = strings.TrimPrefix(container.Names[0], "/")
		data.Image = container.Image
		data.ImageID = container.ImageID
		data.Command = container.Command
		data.Created = time.Unix(container.Created, 0).Local()
		data.State = container.State
		data.Status = container.Status
		data.Port = []Port{}
		for _, p := range container.Ports {
			data.Port = append(data.Port, Port{
				IP:          p.IP,
				PrivatePort: p.PrivatePort,
				PublicPort:  p.PublicPort,
				Type:        p.Type,
			})
		}
		data.SizeRw = container.SizeRw
		data.SizeRootFs = container.SizeRootFs
		data.NetworkMode = container.HostConfig.NetworkMode
		data.NetworkSettings = []string{}
		for _, n := range container.NetworkSettings.Networks {
			data.NetworkSettings = append(data.NetworkSettings, n.NetworkID)
		}
		data.Mounts = ""
		for _, m := range container.Mounts {
			data.Mounts += fmt.Sprintf("%s:%s,", m.Source, m.Destination)
		}
		data.Mounts = strings.TrimSuffix(data.Mounts, ",")
		data.Labels, _ = json.Marshal(container.Labels)

		containerList = append(containerList, data)
	}

	c.JSON(http.StatusOK, gin.H{"containerList": containerList})
}

func GetStats(c *gin.Context) {
	containerID := c.Param("containerID")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is empty"})
		return
	}

	stats, err := utils.GetContainerStatsByID(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cpuUsage := float64(stats.CPUStats.CPUUsage.TotalUsage) / float64(stats.CPUStats.SystemUsage) * 100
	memoryUsage := float64(stats.MemoryStats.Usage) / (1024 * 1024)
	memoryLimit := float64(stats.MemoryStats.Limit) / (1024 * 1024)
	networkRxBytes := float64(stats.Networks["eth0"].RxBytes) / (1024 * 1024)
	networkTxBytes := float64(stats.Networks["eth0"].TxBytes) / (1024 * 1024)

	containerStats := struct {
		ContainerID string `json:"containerID"`
		CPU         string `json:"cpu"`
		Memory      string `json:"memory"`
		NetworkRx   string `json:"networkRx"`
		NetworkTx   string `json:"networkTx"`
	}{
		ContainerID: containerID,
		CPU:         fmt.Sprintf("%.2f%%", cpuUsage),
		Memory:      fmt.Sprintf("%.2f / %.2f MB", memoryUsage, memoryLimit),
		NetworkRx:   fmt.Sprintf("%.2f MB", networkRxBytes),
		NetworkTx:   fmt.Sprintf("%.2f MB", networkTxBytes),
	}

	c.JSON(http.StatusOK, gin.H{"stats": containerStats})
}

func StartContainer(c *gin.Context) {
	containerID := c.Param("containerID")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is empty"})
		return
	}

	err := utils.StartContainerByID(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func StopContainer(c *gin.Context) {
	containerID := c.Param("containerID")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is empty"})
		return
	}

	err := utils.StopContainerByID(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func UpdateContainer(c *gin.Context) {
	containerID := c.Param("containerID")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is empty"})
		return
	}

	err := utils.UpdateContainerByID(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

func DeleteContainer(c *gin.Context) {
	containerID := c.Param("containerID")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "containerID is empty"})
		return
	}

	err := utils.DeleteContainerByID(containerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
