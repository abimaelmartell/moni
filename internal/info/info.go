package info

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Info struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Uptime   string `json:"uptime"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	CPU      string `json:"cpu"`
	Memory   string `json:"memory"`
	Disk     string `json:"disk"`
}

var startTime = time.Now()

func getPrimaryIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "unknown"
	}
	for _, iface := range ifaces {
		// skip down or loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an IPv4 address
			}
			return ip.String()
		}
	}
	return "unknown"
}

func Handler() gin.HandlerFunc {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	ip := getPrimaryIP()

	return func(c *gin.Context) {
		uptime := time.Since(startTime).Truncate(time.Second).String()
		cpuCount := runtime.NumCPU()

		vm, _ := mem.VirtualMemory()
		memUsedGB := float64(vm.Used) / (1024 * 1024 * 1024)
		memTotalGB := float64(vm.Total) / (1024 * 1024 * 1024)

		diskStat, _ := disk.Usage("/")
		diskUsedGB := float64(diskStat.Used) / (1024 * 1024 * 1024)
		diskTotalGB := float64(diskStat.Total) / (1024 * 1024 * 1024)

		info := Info{
			Hostname: hostname,
			IP:       ip,
			Uptime:   uptime,
			OS:       runtime.GOOS,
			Arch:     runtime.GOARCH,
			CPU:      fmt.Sprintf("%d cores", cpuCount),
			Memory:   fmt.Sprintf("%.1f GB / %.1f GB", memUsedGB, memTotalGB),
			Disk:     fmt.Sprintf("%.1f GB / %.1f GB", diskUsedGB, diskTotalGB),
		}

		c.JSON(http.StatusOK, info)
	}
}
