package metrics

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/process"
	"go.etcd.io/bbolt"
)

type ProcInfo struct {
	PID     int32   `json:"pid"`
	CPU     float64 `json:"cpu"`
	Memory  uint64  `json:"memory"`
	Command string  `json:"command"`
}

func GetTopProcesses(sortBy string, limit int) ([]ProcInfo, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var infos []ProcInfo
	for _, p := range procs {
		var (
			cpuPct  float64
			memInfo *process.MemoryInfoStat
			cpuErr  error
			memErr  error
		)

		if sortBy == "cpu" {
			cpuPct, cpuErr = p.CPUPercent()
			if cpuErr != nil {
				continue
			}
			memInfo, memErr = p.MemoryInfo()
			if memErr != nil {
				memInfo = &process.MemoryInfoStat{RSS: 0}
			}
		} else {
			memInfo, memErr = p.MemoryInfo()
			if memErr != nil {
				continue
			}
			cpuPct, cpuErr = p.CPUPercent()
			if cpuErr != nil {
				cpuPct = 0
			}
		}

		name, _ := p.Name()
		infos = append(infos, ProcInfo{
			PID:     p.Pid,
			CPU:     cpuPct,
			Memory:  memInfo.RSS,
			Command: name,
		})
	}

	if sortBy == "memory" {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].Memory > infos[j].Memory
		})
	} else {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].CPU > infos[j].CPU
		})
	}

	if len(infos) > limit {
		infos = infos[:limit]
	}

	return infos, nil
}

type DataPoint struct {
	Timestamp  int64   `json:"timestamp"`
	CPUPercent float64 `json:"cpu_percent"`
	MemTotal   uint64  `json:"mem_total"`
	MemUsed    uint64  `json:"mem_used"`
	DiskTotal  uint64  `json:"disk_total"`
	DiskUsed   uint64  `json:"disk_used"`
	LoadAvg    struct {
		Load1  float64 `json:"load1"`
		Load5  float64 `json:"load5"`
		Load15 float64 `json:"load15"`
	} `json:"load_avg"`
	MemPercent  float64 `json:"mem_percent"`
	DiskPercent float64 `json:"disk_percent"`
}

type MetricsResponse struct {
	DataPoints   []DataPoint `json:"data_points"`
	TopProcesses []ProcInfo  `json:"top_processes"`
}

func Handler(db *bbolt.DB, historyLimit, procLimit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		sortBy := c.DefaultQuery("sortProcessesBy", "cpu")

		if sortBy != "cpu" && sortBy != "memory" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sort parameter"})
			return
		}

		var points []DataPoint
		if err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("Metrics"))
			if b == nil {
				return nil
			}
			cur := b.Cursor()
			for k, v := cur.Last(); k != nil && len(points) < historyLimit; k, v = cur.Prev() {
				var p DataPoint
				if err := json.Unmarshal(v, &p); err != nil {
					return err
				}
				points = append(points, p)
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}

		procs, err := GetTopProcesses(sortBy, procLimit)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

			return
		}

		response := MetricsResponse{
			DataPoints:   points,
			TopProcesses: procs,
		}

		c.JSON(http.StatusOK, response)
	}
}
