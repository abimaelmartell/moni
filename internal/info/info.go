package info

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
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

	OS              string `json:"os"`               // e.g. "linux"
	Platform        string `json:"platform"`         // e.g. "ubuntu"
	PlatformVersion string `json:"platform_version"` // e.g. "22.04"
	KernelVersion   string `json:"kernel_version"`   // e.g. "5.15.0-46-generic"

	Uptime   string `json:"uptime"`    // real system uptime
	CPUModel string `json:"cpu_model"` // e.g. "Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz"
	CPUCores int    `json:"cpu_cores"` // number of logical cores

	Memory string `json:"memory"` // used / total
	Swap   string `json:"swap"`   // used / total swap
	Disk   string `json:"disk"`   // used / total on “/”

	UpdateInterval int `json:"update_interval"` // how often it refreshes
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

func Handler(updateInterval int) gin.HandlerFunc {
	// fetch host info once
	hi, _ := host.Info()
	ci, _ := cpu.Info()

	// determine primary interface & IP …
	ifaceName, ip := getPrimaryInterface()

	return func(c *gin.Context) {
		// 1) uptime
		uSecs, _ := host.Uptime()
		uptime := time.Duration(uSecs) * time.Second

		// 2) CPU cores
		cores := runtime.NumCPU()

		// 3) mem + swap
		vm, _ := mem.VirtualMemory()
		sm, _ := mem.SwapMemory()

		// 4) disk
		ds, _ := disk.Usage("/")

		info := Info{
			Hostname:  hi.Hostname,
			Interface: ifaceName,
			IP:        ip,

			OS:              hi.OS,
			Platform:        hi.Platform,
			PlatformVersion: hi.PlatformVersion,
			KernelVersion:   hi.KernelVersion,

			Uptime:   uptime.Truncate(time.Second).String(),
			CPUModel: ci[0].ModelName,
			CPUCores: cores,

			Memory: fmt.Sprintf("%.1fGB/%.1fGB", float64(vm.Used)/(1<<30), float64(vm.Total)/(1<<30)),
			Swap:   fmt.Sprintf("%.1fGB/%.1fGB", float64(sm.Used)/(1<<30), float64(sm.Total)/(1<<30)),
			Disk:   fmt.Sprintf("%.1fGB/%.1fGB", float64(ds.Used)/(1<<30), float64(ds.Total)/(1<<30)),

			UpdateInterval: updateInterval,
		}

		c.JSON(http.StatusOK, info)
	}
}
