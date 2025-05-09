package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/abimaelmartell/moni/internal/collector"
	"github.com/abimaelmartell/moni/internal/info"
	"github.com/abimaelmartell/moni/internal/metrics"
	"github.com/gin-gonic/gin"
)

//go:embed static
var embeddedFS embed.FS

func main() {
	staticFS, err := fs.Sub(embeddedFS, "static")
	if err != nil {
		log.Fatalf("failed to sub FS: %v", err)
	}

	port := os.Getenv("MONI_PORT")
	if port == "" {
		port = "8080"
	}

	metricsPath := os.Getenv("MONI_METRICS_PATH")
	if metricsPath == "" {
		metricsPath = "./metrics.db"
	}

	updateInterval := os.Getenv("MONI_UPDATE_INTERVAL")
	if updateInterval == "" {
		updateInterval = "1s"
	}

	interval, err := time.ParseDuration(updateInterval)
	if err != nil {
		log.Fatalf("invalid update interval: %v", err)
	}

	db := collector.OpenBolt(metricsPath)
	defer db.Close()
	go collector.Run(db, interval)

	r := gin.Default()

	r.GET("/metrics", metrics.Handler(db, int(interval.Milliseconds()), 10))
	r.GET("/info", info.Handler(int(interval.Milliseconds())))

	r.StaticFS("/static", http.FS(staticFS))

	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	log.Printf("Starting server on :%s", port)

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
