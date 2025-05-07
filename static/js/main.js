document.addEventListener('DOMContentLoaded', () => {
  const cpu_ctx = document
    .getElementById('cpuChart')
    .getContext('2d')

  const mem_ctx = document
    .getElementById('memChart')
    .getContext('2d')

  const load_ctx = document
    .getElementById('loadChart')
    .getContext('2d')

  const cpu_chart = new Chart(cpu_ctx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [{
        label: 'CPU %',
        data: [],
        borderColor: '#3b82f6',
        fill: false
      }]
    },
    options: {
      responsive: false,
      animation: false,
      scales: { y: { beginAtZero: true } }
    }
  })

  const mem_chart = new Chart(mem_ctx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [{
        label: 'Memory %',
        data: [],
        borderColor: '#3b82f6',
        fill: false
      }]
    },
    options: {
      responsive: false,
      animation: false,
      scales: { y: { beginAtZero: true } }
    }
  })

  const load_chart = new Chart(load_ctx, {
    type: 'line',
    data: {
      labels: [],
      datasets: [{
        label: 'Load Average',
        data: [],
        borderColor: '#3b82f6',
        fill: false
      }]
    },
    options: {
      responsive: false,
      animation: false,
      scales: { y: { beginAtZero: true } }
    }
  })

  async function fetchAndUpdate() {
    try {
      const res = await fetch('/metrics')
      if (!res.ok) throw new Error(res.status)
      const points = await res.json()
      if (!Array.isArray(points) || !points.length) return

      cpu_chart.data.labels = points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      cpu_chart.data.datasets[0].data = points.map(p =>
        p.cpu_percent
      )

      mem_chart.data.labels = points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      mem_chart.data.datasets[0].data = points.map(p =>
        p.mem_percent
      )

      load_chart.data.labels = points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      load_chart.data.datasets[0].data = points.map(p =>
        p.load_avg.load1
      )


      cpu_chart.update()
      mem_chart.update()
      load_chart.update()

      const topProcessesBody = document.getElementById('topProcessesBody')
      topProcessesBody.innerHTML = ''

      const last = points[points.length - 1]

      document.getElementById('cpuVal').textContent = `${last.cpu_percent.toFixed(1)}%`
      document.getElementById('memVal').textContent = `${((last.mem_used / last.mem_total) * 100).toFixed(1)}%`
      document.getElementById('loadVal').textContent =
        `${last.load_avg.load1.toFixed(1)} / ${last.load_avg.load5.toFixed(1)} / ${last.load_avg.load15.toFixed(1)}`
      document.getElementById('diskVal').textContent =
        `${((last.disk_used / last.disk_total) * 100).toFixed(1)}%`


      last.top_processes.forEach(p => {
        const row = document.createElement('tr')
        const memoryMB = parseInt(p.memory / 1024 / 1024)
        row.innerHTML = `
            <td class="border border-gray-300 p-2 text-right">${p.pid}</td>
            <td class="border border-gray-300 p-2 text-right">${p.cpu.toFixed(2)}%</td>
            <td class="border border-gray-300 p-2 text-right">${memoryMB} MB</td>
            <td class="border border-gray-300 p-2">${p.command}</td>
          `
        topProcessesBody.appendChild(row)
      })
    } catch (e) {
      console.error('update error', e)
    }
  }

  async function fetchAndUpdateInfo() {
    try {
      const res = await fetch('/info')
      if (!res.ok) throw new Error(res.status)
      const info = await res.json()

      document.getElementById('hostname').textContent = info.hostname
      document.getElementById('hostname_info').textContent = info.hostname
      document.getElementById('ip').textContent = info.ip
      document.getElementById('uptime').textContent = info.uptime
      document.getElementById('os').textContent = info.os
      document.getElementById('arch').textContent = info.arch
      document.getElementById('memory').textContent = info.memory
      document.getElementById('disk').textContent = info.disk
    } catch (e) {
      console.error('update info error', e)
    }
  }

  fetchAndUpdate()
  fetchAndUpdateInfo()
  setInterval(fetchAndUpdate, 1000)
})

document.getElementById('openInfo').onclick = () => {
  document.getElementById('infoModal').classList.remove('hidden')
}
document.getElementById('closeInfo').onclick = () => {
  document.getElementById('infoModal').classList.add('hidden')
}
