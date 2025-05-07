package metrics

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"
)

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
	MemPercent   float64 `json:"mem_percent"`
	DiskPercent  float64 `json:"disk_percent"`
	TopProcesses []struct {
		PID     int     `json:"pid"`
		CPU     float64 `json:"cpu"`
		Memory  uint64  `json:"memory"`
		Command string  `json:"command"`
	} `json:"top_processes"`
	Hostname string `json:"hostname"`
}

func Handler(db *bbolt.DB, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var points []DataPoint

		err := db.View(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("Metrics"))
			if b == nil {
				return nil
			}
			cursor := b.Cursor()

			k, v := cursor.Last()
			count := 0
			for k != nil && count < limit {
				var p DataPoint
				if err := json.Unmarshal(v, &p); err != nil {
					return err
				}
				points = append(points, p)
				k, v = cursor.Prev()
				count++
			}
			return nil
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}

		c.JSON(http.StatusOK, points)
	}
}
