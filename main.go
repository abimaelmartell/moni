package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/abimaelmartell/moni/internal/collector"
	"github.com/abimaelmartell/moni/internal/metrics"
)

func main() {
	db := collector.OpenBolt("./metrics.db")
	defer db.Close()

	go collector.Run(db, time.Second)

	r := gin.Default()

	r.GET("/metrics", metrics.Handler(db, 60))

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	fmt.Println("Starting server on :8080")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
