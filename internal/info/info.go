package info

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type Info struct {
	Hostname  string `json:"hostname"`
	Interface string `json:"interface"`
	IP        string `json:"ip"`

	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`

	Uptime   string `json:"uptime"`
	CPUModel string `json:"cpu_model"`
	CPUCores int    `json:"cpu_cores"`

	TotalMemory int `json:"total_memory"`
	UsedMemory  int `json:"used_memory"`
	TotalSwap   int `json:"total_swap"`
	UsedSwap    int `json:"used_swap"`
	TotalDisk   int `json:"total_disk"`
	UsedDisk    int `json:"used_disk"`

	UpdateInterval int `json:"update_interval"`
}

func getPrimaryInterface() (string, string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown", "unknown"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.IsLoopback() {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}
			return iface.Name, ip4.String()
		}
	}
	return "unknown", "unknown"
}

func getFormattedUptime() (string, error) {
	uSecs, err := host.Uptime()
	if err != nil {
		return "", err
	}
	upt := time.Duration(uSecs) * time.Second

	days := upt / (24 * time.Hour)
	hours := (upt % (24 * time.Hour)) / time.Hour
	minutes := (upt % time.Hour) / time.Minute
	seconds := (upt % time.Minute) / time.Second

	parts := []string{}
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%d days", days))
	}
	parts = append(parts, fmt.Sprintf("%d hours", hours))
	parts = append(parts, fmt.Sprintf("%d minutes", minutes))
	parts = append(parts, fmt.Sprintf("%d seconds", seconds))

	formatted := strings.Join(parts, ", ")
	return formatted, nil
}

func Handler(updateInterval int) gin.HandlerFunc {
	hi, _ := host.Info()
	ci, _ := cpu.Info()

	ifaceName, ip := getPrimaryInterface()

	return func(c *gin.Context) {
		uptime, err := getFormattedUptime()

		if err != nil {
			uptime = "unknown"
		}

		cores := runtime.NumCPU()

		vm, _ := mem.VirtualMemory()
		sm, _ := mem.SwapMemory()

		ds, _ := disk.Usage("/")

		info := Info{
			Hostname:  hi.Hostname,
			Interface: ifaceName,
			IP:        ip,

			OS:              hi.OS,
			Platform:        hi.Platform,
			PlatformVersion: hi.PlatformVersion,
			KernelVersion:   hi.KernelVersion,

			Uptime:   uptime,
			CPUModel: ci[0].ModelName,
			CPUCores: cores,

			TotalMemory: int(vm.Total),
			UsedMemory:  int(vm.Used),
			TotalSwap:   int(sm.Total),
			UsedSwap:    int(sm.Used),
			TotalDisk:   int(ds.Total),
			UsedDisk:    int(ds.Used),

			UpdateInterval: updateInterval,
		}

		c.JSON(http.StatusOK, info)
	}
}
