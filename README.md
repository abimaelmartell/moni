# moni

> Ultra-lightweight system metrics web dashboard in Go

A single static binary that collects CPU, memory and disk usage samples, stores them in an embedded BoltDB, and serves both a JSON API and a plain-HTML/Chart.js dashboard—no build tools required.

## 🚀 Quick Start


```bash
git clone https://github.com/abimaelmartell/moni.git
cd moni

# Download dependencies
go mod tidy

# Run Directly from Code (no build for now)
go run main.go
```
