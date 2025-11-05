<template>
  <div class="min-h-screen bg-gray-100">
    <!-- Header -->
    <header class="bg-white shadow">
      <div class="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8 flex justify-between items-center">
        <div>
          <h1 class="text-2xl font-bold text-gray-900">IoT Temperature Monitor</h1>
          <p class="text-sm text-gray-600">Blockchain-based real-time monitoring</p>
        </div>
        <div class="flex items-center gap-4">
          <span class="text-sm text-gray-600">{{ user?.username }} ({{ user?.role }})</span>
          <button
            @click="handleLogout"
            class="px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700 transition duration-200"
          >
            Logout
          </button>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <!-- Stats Overview -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center">
            <div class="flex-1">
              <p class="text-sm text-gray-600">Total Devices</p>
              <p class="text-3xl font-bold text-gray-900">{{ devices.length }}</p>
            </div>
            <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z" />
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center">
            <div class="flex-1">
              <p class="text-sm text-gray-600">Active Sensors</p>
              <p class="text-3xl font-bold text-green-600">{{ activeSensors }}</p>
            </div>
            <div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
        </div>

        <div class="bg-white rounded-lg shadow p-6">
          <div class="flex items-center">
            <div class="flex-1">
              <p class="text-sm text-gray-600">Avg Temperature</p>
              <p class="text-3xl font-bold text-orange-600">{{ averageTemp }}°C</p>
            </div>
            <div class="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 6l3 1m0 0l-3 9a5.002 5.002 0 006.001 0M6 7l3 9M6 7l6-2m6 2l3-1m-3 1l-3 9a5.002 5.002 0 006.001 0M18 7l3 9m-3-9l-6-2m0-2v2m0 16V5m0 16H9m3 0h3" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Refresh Button -->
      <div class="mb-4 flex justify-end">
        <button
          @click="loadDevices"
          :disabled="loading"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition duration-200 disabled:opacity-50"
        >
          {{ loading ? 'Refreshing...' : 'Refresh Data' }}
        </button>
      </div>

      <!-- Device Cards Grid -->
      <div v-if="loading && devices.length === 0" class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <p class="mt-4 text-gray-600">Loading devices...</p>
      </div>

      <div v-else-if="devices.length === 0" class="text-center py-12 bg-white rounded-lg shadow">
        <p class="text-gray-600">No devices found. Please wait for devices to register.</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <DeviceCard
          v-for="device in devices"
          :key="device.deviceID"
          :device="device"
          @refresh="loadDevices"
        />
      </div>
    </main>
  </div>
</template>

<script>
import api from '../utils/api'
import DeviceCard from '../components/DeviceCard.vue'

export default {
  name: 'Dashboard',
  components: {
    DeviceCard
  },
  data() {
    return {
      user: null,
      devices: [],
      loading: false,
      refreshInterval: null
    }
  },
  computed: {
    activeSensors() {
      return this.devices.filter(d => d.status === 'active').length
    },
    averageTemp() {
      if (this.devices.length === 0) return 0
      const temps = this.devices
        .filter(d => d.lastReading)
        .map(d => d.lastReading.temperature)
      if (temps.length === 0) return 0
      const avg = temps.reduce((sum, t) => sum + t, 0) / temps.length
      return avg.toFixed(1)
    }
  },
  mounted() {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      this.user = JSON.parse(userStr)
    }
    this.loadDevices()

    // Auto-refresh every 5 seconds
    this.refreshInterval = setInterval(() => {
      this.loadDevices(true)
    }, 5000)
  },
  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
    }
  },
  methods: {
    async loadDevices(silent = false) {
      if (!silent) this.loading = true

      try {
        const response = await api.get('/devices')
        this.devices = response.data.devices || []
      } catch (error) {
        console.error('Failed to load devices:', error)
      } finally {
        this.loading = false
      }
    },

    handleLogout() {
      localStorage.removeItem('authToken')
      localStorage.removeItem('user')
      this.$router.push('/login')
    }
  }
}
</script>
