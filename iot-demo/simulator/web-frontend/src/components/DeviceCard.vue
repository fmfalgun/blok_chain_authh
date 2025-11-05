<template>
  <div class="bg-white rounded-lg shadow-lg p-6 hover:shadow-xl transition duration-200">
    <!-- Device Header -->
    <div class="flex justify-between items-start mb-4">
      <div>
        <h3 class="text-lg font-semibold text-gray-900">{{ device.deviceName }}</h3>
        <p class="text-sm text-gray-500">{{ device.deviceID }}</p>
      </div>
      <span :class="statusClass" class="px-2 py-1 text-xs rounded-full">
        {{ device.status }}
      </span>
    </div>

    <!-- Current Temperature -->
    <div v-if="device.lastReading" class="text-center mb-4">
      <div class="relative">
        <div :class="tempColorClass" class="text-5xl font-bold py-4 rounded-lg">
          {{ device.lastReading.temperature.toFixed(1) }}°C
        </div>
        <div v-if="isAnomaly" class="absolute top-2 right-2">
          <svg class="w-6 h-6 text-red-600 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        </div>
      </div>
      <p class="text-xs text-gray-500 mt-2">
        Last updated: {{ formatTime(device.lastReading.timestamp) }}
      </p>
    </div>

    <div v-else class="text-center py-8 text-gray-400">
      <svg class="w-12 h-12 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
      </svg>
      <p>No data yet</p>
    </div>

    <!-- Mini Chart -->
    <div v-if="showChart && chartData" class="mb-4">
      <canvas ref="chartCanvas" class="w-full h-32"></canvas>
    </div>

    <!-- Actions -->
    <div class="flex gap-2">
      <button
        @click="viewDetails"
        class="flex-1 px-3 py-2 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition duration-200"
      >
        View Details
      </button>
      <button
        @click="$emit('refresh')"
        class="px-3 py-2 bg-gray-200 text-gray-700 text-sm rounded hover:bg-gray-300 transition duration-200"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>
    </div>

    <!-- Details Modal (simplified) -->
    <div v-if="showDetails" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" @click="showDetails = false">
      <div class="bg-white rounded-lg p-6 max-w-2xl w-full m-4" @click.stop>
        <h2 class="text-2xl font-bold mb-4">{{ device.deviceName }}</h2>
        <div class="grid grid-cols-2 gap-4 mb-4">
          <div>
            <p class="text-sm text-gray-600">Device ID</p>
            <p class="font-semibold">{{ device.deviceID }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Type</p>
            <p class="font-semibold">{{ device.deviceType }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Status</p>
            <p class="font-semibold">{{ device.status }}</p>
          </div>
          <div>
            <p class="text-sm text-gray-600">Owner</p>
            <p class="font-semibold">{{ device.ownerID }}</p>
          </div>
        </div>
        <button
          @click="showDetails = false"
          class="w-full px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700 transition duration-200"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import { Chart, registerables } from 'chart.js'
Chart.register(...registerables)

export default {
  name: 'DeviceCard',
  props: {
    device: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      showDetails: false,
      showChart: false,
      chartData: null,
      chart: null
    }
  },
  computed: {
    statusClass() {
      return this.device.status === 'active'
        ? 'bg-green-100 text-green-800'
        : 'bg-gray-100 text-gray-800'
    },
    tempColorClass() {
      if (!this.device.lastReading) return 'text-gray-400'

      const temp = this.device.lastReading.temperature
      if (temp < 20) return 'text-blue-600 bg-blue-50'
      if (temp >= 20 && temp <= 25) return 'text-green-600 bg-green-50'
      if (temp > 25 && temp <= 28) return 'text-orange-500 bg-orange-50'
      return 'text-red-600 bg-red-50'
    },
    isAnomaly() {
      if (!this.device.lastReading) return false
      const temp = this.device.lastReading.temperature
      return temp < 18 || temp > 28
    }
  },
  methods: {
    formatTime(timestamp) {
      const date = new Date(timestamp * 1000)
      return date.toLocaleTimeString()
    },
    viewDetails() {
      this.showDetails = true
    }
  }
}
</script>
