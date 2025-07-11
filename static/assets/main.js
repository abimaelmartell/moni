let updateIntervalMs = 1000
let sortBy = localStorage.getItem('sortBy') || 'cpu'

let isLoadingProcs = true

const loadingTbody = document.getElementById('topProcessesLoading')
const dataTbody = document.getElementById('topProcessesBody')

const fixTableHeaders = () => {
  if (sortBy === 'memory') {
    document.querySelector('#topProcessesTable thead th:nth-child(2)').textContent = 'Memory'
    document.querySelector('#topProcessesTable thead th:nth-child(3)').textContent = 'CPU'
  } else {
    document.querySelector('#topProcessesTable thead th:nth-child(2)').textContent = 'CPU'
    document.querySelector('#topProcessesTable thead th:nth-child(3)').textContent = 'Memory'
  }
}

const infoModal = document.getElementById('infoModal')

const openInfoModal = () => {
  infoModal.classList.remove('hidden')

  requestAnimationFrame(() => {
    infoModal.classList.remove('opacity-0')
    infoModal.classList.add('opacity-100')
  })
}

const closeInfoModal = () => {
  infoModal.classList.remove('opacity-100')
  infoModal.classList.add('opacity-0')

  infoModal.addEventListener('transitionend', function listener() {
    infoModal.classList.add('hidden')
    infoModal.removeEventListener('transitionend', listener)
  })
}

document.getElementById('openInfo').onclick = () => {
  openInfoModal()
}

document.getElementById('closeInfo').onclick = () => {
  closeInfoModal()
}

document.getElementById('infoModal').onclick = (e) => {
  if (e.target === document.getElementById('infoModal')) {
    closeInfoModal()
  }
}

document.addEventListener('keydown', (e) => {
  if (e.key === 'Escape') {
    closeInfoModal()
  }
})

const formatBytes = (bytes) => {
  const fmt = (num, dec = 2) =>
    num.toLocaleString(undefined, {
      minimumFractionDigits: dec,
      maximumFractionDigits: dec,
    });

  if (bytes < 1024) {
    return bytes.toLocaleString() + ' B'
  }

  if (bytes < 1024 ** 2) {
    return fmt(bytes / 1024) + ' KB'
  }

  if (bytes < 1024 ** 3) {
    return fmt(bytes / 1024 ** 2) + ' MB'
  }

  return fmt(bytes / 1024 ** 3) + ' GB'
}

const formatOS = (os, platform, platform_version, kernel_version) => {
  if (os === 'linux') {
    const capitalizedPlatform = platform.charAt(0).toUpperCase() + platform.slice(1)
    return capitalizedPlatform + ' ' + platform_version + ' (kernel ' + kernel_version + ')'
  }

  if (os === 'darwin') {
    return 'MacOS' + ' ' + platform_version + '  (kernel ' + kernel_version + ')'
  }

  return os + ' ' + platform + ' ' + platform_version + ' ' + kernel_version
}

document.getElementById('sortSelect').addEventListener('change', e => {
  sortBy = e.target.value
  localStorage.setItem('sortBy', sortBy)

  isLoadingProcs = true
  dataTbody.innerHTML = ''

  loadingTbody.classList.remove('hidden')
  dataTbody.classList.add('hidden')

  fixTableHeaders()
})

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
      const res = await fetch('/metrics?sortProcessesBy=' + sortBy)
      if (!res.ok) throw new Error(res.status)
      const data = await res.json()
      if (!Array.isArray(data.data_points) || !data.data_points.length) return

      cpu_chart.data.labels = data.data_points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      cpu_chart.data.datasets[0].data = data.data_points.map(p =>
        p.cpu_percent
      )

      mem_chart.data.labels = data.data_points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      mem_chart.data.datasets[0].data = data.data_points.map(p =>
        p.mem_percent
      )

      load_chart.data.labels = data.data_points.map(p =>
        new Date(p.timestamp * 1000).toLocaleTimeString()
      )
      load_chart.data.datasets[0].data = data.data_points.map(p =>
        p.load_avg.load1
      )

      cpu_chart.update()
      mem_chart.update()
      load_chart.update()

      const topProcessesBody = document.getElementById('topProcessesBody')
      topProcessesBody.innerHTML = ''

      const last = data.data_points[data.data_points.length - 1]

      document.getElementById('cpuVal').textContent = `${last.cpu_percent.toFixed(1)}%`
      document.getElementById('memVal').textContent = `${((last.mem_used / last.mem_total) * 100).toFixed(1)}%`

      const memUsedBytes = formatBytes(last.mem_used)
      const memTotalBytes = formatBytes(last.mem_total)
      const diskUsedBytes = formatBytes(last.disk_used)
      const diskTotalBytes = formatBytes(last.disk_total)

      document.getElementById('memDetail').textContent =
        `${memUsedBytes} used of ${memTotalBytes}`

      document.getElementById('diskDetail').textContent =
        `${diskUsedBytes} used of ${diskTotalBytes}`

      document.getElementById('loadVal').textContent =
        `${last.load_avg.load1.toFixed(1)} / ${last.load_avg.load5.toFixed(1)} / ${last.load_avg.load15.toFixed(1)}`

      document.getElementById('diskVal').textContent =
        `${((last.disk_used / last.disk_total) * 100).toFixed(1)}%`

      data.top_processes.forEach(p => {
        const row = document.createElement('tr')
        row.classList.add('hover:bg-gray-100')

        const memoryMB = (p.memory / 1024 / 1024).toFixed(0)

        let metricCells

        if (sortBy === 'memory') {
          metricCells = `
              <td class="border border-gray-300 p-2 text-right">${memoryMB} MB</td>
              <td class="border border-gray-300 p-2 text-right">${p.cpu.toFixed(2)}%</td>
            `
        } else {
          metricCells = `
              <td class="border border-gray-300 p-2 text-right">${p.cpu.toFixed(2)}%</td>
              <td class="border border-gray-300 p-2 text-right">${memoryMB} MB</td>
            `
        }

        row.innerHTML = `
            <td class="border border-gray-300 p-2 text-right">${p.pid}</td>
            ${metricCells}
            <td class="border border-gray-300 p-2">${p.command}</td>
          `
        topProcessesBody.appendChild(row)
      })

      if (isLoadingProcs) {
        loadingTbody.classList.add('hidden')
        dataTbody.classList.remove('hidden')
        isLoadingProcs = false
      }
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
      document.getElementById('os').textContent = formatOS(info.os, info.platform, info.platform_version, info.kernel_version)
      document.getElementById('cpu').textContent = info.cpu_model + ' (' + info.cpu_cores + ' cores)'
      document.getElementById('memory').textContent = formatBytes(info.used_memory) + ' / ' + formatBytes(info.total_memory)
      document.getElementById('disk').textContent = formatBytes(info.used_disk) + ' / ' + formatBytes(info.total_disk)

      updateIntervalMs = info.update_interval

      document.getElementById('updateInterval').textContent = `${updateIntervalMs / 1000}s`
      document.getElementById('interface').textContent = info.interface
    } catch (e) {
      console.error('update info error', e)
    }
  }

  fetchAndUpdate()
  fetchAndUpdateInfo()

  setInterval(fetchAndUpdate, updateIntervalMs)

  fixTableHeaders()
})
