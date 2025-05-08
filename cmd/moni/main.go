package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/abimaelmartell/moni/internal/collector"
	"github.com/abimaelmartell/moni/internal/info"
	"github.com/abimaelmartell/moni/internal/metrics"
)

func main() {
	metrics_path := os.Getenv("MONI_METRICS_PATH")

	if metrics_path == "" {
		metrics_path = "./metrics.db"
	}

	fmt.Println("metrics_path", metrics_path)

	update_interval := os.Getenv("MONI_UPDATE_INTERVAL")

	if update_interval == "" {
		update_interval = "1s"
	}

	interval, err := time.ParseDuration(update_interval)
	if err != nil {
		log.Fatalf("invalid update interval: %v", err)
	}

	db := collector.OpenBolt(metrics_path)
	defer db.Close()

	go collector.Run(db, interval)

	r := gin.Default()

	r.GET("/metrics", metrics.Handler(db, int(interval.Milliseconds())))
	r.GET("/info", info.Handler(int(interval.Milliseconds())))

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	fmt.Println("Starting server on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
