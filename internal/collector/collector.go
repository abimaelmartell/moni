package collector

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"time"

	"go.etcd.io/bbolt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

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

		record, err := json.Marshal(struct {
			Timestamp  int64   `json:"timestamp"`
			CPUPercent float64 `json:"cpu_percent"`
			MemTotal   uint64  `json:"mem_total"`
			MemUsed    uint64  `json:"mem_used"`
			DiskTotal  uint64  `json:"disk_total"`
			DiskUsed   uint64  `json:"disk_used"`
		}{
			Timestamp:  now.Unix(),
			CPUPercent: cpuPercents[0],
			MemTotal:   vmStat.Total,
			MemUsed:    vmStat.Used,
			DiskTotal:  diskStat.Total,
			DiskUsed:   diskStat.Used,
		})
		if err != nil {
			log.Printf("warning: failed to marshal record: %v", err)
			continue
		}

		err = db.Update(func(tx *bbolt.Tx) error {
			b := tx.Bucket([]byte("Metrics"))
			return b.Put(itob(now.Unix()), record)
		})
		if err != nil {
			log.Printf("warning: failed to write record to db: %v", err)
		}
	}
}
