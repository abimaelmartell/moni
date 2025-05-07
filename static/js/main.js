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
        p.load_avg
      )


      cpu_chart.update()
      mem_chart.update()
      load_chart.update()

      const topProcessesBody = document.getElementById('topProcessesBody')
      topProcessesBody.innerHTML = ''

      const lastPoint = points[points.length - 1]

      lastPoint.top_processes.forEach(p => {
        const row = document.createElement('tr')
        const memoryMB = p.memory / 1024 / 1024
        row.innerHTML = `
            <td class="border border-gray-300 p-2">${p.pid}</td>
            <td class="border border-gray-300 p-2">${p.cpu.toFixed(2)}%</td>
            <td class="border border-gray-300 p-2">${memoryMB.toFixed(2)}MB</td>
            <td class="border border-gray-300 p-2">${p.command}</td>
          `
        topProcessesBody.appendChild(row)
      })
    } catch (e) {
      console.error('update error', e)
    }
  }

  fetchAndUpdate()
  setInterval(fetchAndUpdate, 1000)
})