# moni

> Ultra-lightweight system metrics web dashboard in Go

A single static binary that collects CPU, memory and disk usage samples, stores them in an embedded BoltDB, and serves both a JSON API and a plain-HTML/Chart.js dashboard—no build tools required.

<img width="1436" alt="Screenshot of Web UI" src="https://github.com/user-attachments/assets/375adfee-1e27-40f2-a09d-165df7817f52" />


## 🚀 Quick Start


```bash
git clone https://github.com/abimaelmartell/moni.git
cd moni

# Download dependencies
go mod tidy

# Run Directly from Code (no build for now)
go run main.go
```
