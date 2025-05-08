package collector

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"go.etcd.io/bbolt"
)

type DataPoint struct {
	Timestamp  int64   `json:"timestamp"`
	CPUPercent float64 `json:"cpu_percent"`
	MemTotal   uint64  `json:"mem_total"`
	MemUsed    uint64  `json:"mem_used"`
	LoadAvg    struct {
		Load1  float64 `json:"load1"`
		Load5  float64 `json:"load5"`
		Load15 float64 `json:"load15"`
	} `json:"load_avg"`
	MemPercent  float64 `json:"mem_percent"`
	DiskTotal   uint64  `json:"disk_total"`
	DiskUsed    uint64  `json:"disk_used"`
	DiskPercent float64 `json:"disk_percent"`
}

func OpenBolt(path string) *bbolt.DB {
	db, err := bbolt.Open(path, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("failed to open bolt db: %v", err)
	}
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Metrics"))
		return err
	})
	if err != nil {
		log.Fatalf("failed to create Metrics bucket: %v", err)
	}
	return db
}

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

func Run(db *bbolt.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for now := range ticker.C {
		cpuPercents, err := cpu.Percent(0, false)
		if err != nil || len(cpuPercents) == 0 {
			log.Printf("warning: failed to get CPU percent: %v", err)
			continue
		}
		vmStat, err := mem.VirtualMemory()
		if err != nil {
			log.Printf("warning: failed to get memory stats: %v", err)
			continue
		}
		diskStat, err := disk.Usage("/")
		if err != nil {
			log.Printf("warning: failed to get disk stats: %v", err)
			continue
		}
		loadAvg, err := load.Avg()
		if err != nil {
			log.Printf("warning: failed to get load average: %v", err)
			continue
		}

		record, err := json.Marshal(DataPoint{
			Timestamp:  now.Unix(),
			CPUPercent: cpuPercents[0],
			MemTotal:   vmStat.Total,
			MemUsed:    vmStat.Used,
			DiskTotal:  diskStat.Total,
			DiskUsed:   diskStat.Used,
			LoadAvg: struct {
				Load1  float64 `json:"load1"`
				Load5  float64 `json:"load5"`
				Load15 float64 `json:"load15"`
			}{
				Load1:  loadAvg.Load1,
				Load5:  loadAvg.Load5,
				Load15: loadAvg.Load15,
			},
			MemPercent:  vmStat.UsedPercent,
			DiskPercent: diskStat.UsedPercent,
		})

		if err != nil {
			log.Printf("warning: failed to marshal record: %v", err)
			continue
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("Metrics"))

			key := itob(now.Unix())
			if err := b.Put(key, record); err != nil {
				return err
			}

			stats := b.Stats()
			if stats.KeyN > 100 {
				toRemove := stats.KeyN - 100

				c := b.Cursor()
				k, _ := c.First()
				for i := 0; i < toRemove && k != nil; i++ {
					if err := c.Delete(); err != nil {
						return err
					}
					k, _ = c.Next()
				}
			}

			return nil
		})

		if err != nil {
			log.Printf("warning: failed to write record to db: %v", err)
		}
	}
}
