<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-purple-600">
    <div class="bg-white rounded-lg shadow-2xl p-8 w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-3xl font-bold text-gray-800">IoT Monitor</h1>
        <p class="text-gray-600 mt-2">Blockchain-Based Temperature Monitoring</p>
      </div>

      <!-- Login Form -->
      <div v-if="!showRegister">
        <h2 class="text-2xl font-semibold mb-6 text-gray-800">Login</h2>
        <form @submit.prevent="handleLogin">
          <div class="mb-4">
            <label class="block text-gray-700 text-sm font-bold mb-2">Username</label>
            <input
              v-model="loginForm.username"
              type="text"
              required
              class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter username"
            />
          </div>
          <div class="mb-6">
            <label class="block text-gray-700 text-sm font-bold mb-2">Password</label>
            <input
              v-model="loginForm.password"
              type="password"
              required
              class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter password"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition duration-200 disabled:opacity-50"
          >
            {{ loading ? 'Logging in...' : 'Login' }}
          </button>
        </form>
        <p class="text-center mt-4 text-gray-600">
          Don't have an account?
          <a href="#" @click.prevent="showRegister = true" class="text-blue-600 hover:underline">Register</a>
        </p>
      </div>

      <!-- Register Form -->
      <div v-else>
        <h2 class="text-2xl font-semibold mb-6 text-gray-800">Register</h2>
        <form @submit.prevent="handleRegister">
          <div class="mb-4">
            <label class="block text-gray-700 text-sm font-bold mb-2">Username</label>
            <input
              v-model="registerForm.username"
              type="text"
              required
              class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Choose username"
            />
          </div>
          <div class="mb-4">
            <label class="block text-gray-700 text-sm font-bold mb-2">Email</label>
            <input
              v-model="registerForm.email"
              type="email"
              required
              class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Enter email"
            />
          </div>
          <div class="mb-6">
            <label class="block text-gray-700 text-sm font-bold mb-2">Password</label>
            <input
              v-model="registerForm.password"
              type="password"
              required
              minlength="6"
              class="w-full px-4 py-2 border rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Choose password (min 6 chars)"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full bg-green-600 text-white py-2 rounded-lg hover:bg-green-700 transition duration-200 disabled:opacity-50"
          >
            {{ loading ? 'Registering...' : 'Register' }}
          </button>
        </form>
        <p class="text-center mt-4 text-gray-600">
          Already have an account?
          <a href="#" @click.prevent="showRegister = false" class="text-blue-600 hover:underline">Login</a>
        </p>
      </div>

      <!-- Error Message -->
      <div v-if="error" class="mt-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
        {{ error }}
      </div>

      <!-- Success Message -->
      <div v-if="success" class="mt-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded">
        {{ success }}
      </div>

      <!-- Demo Accounts -->
      <div class="mt-6 p-4 bg-gray-100 rounded-lg text-sm">
        <p class="font-semibold text-gray-700 mb-2">Demo Accounts:</p>
        <p class="text-gray-600">alice / alice123</p>
        <p class="text-gray-600">bob / bob123</p>
        <p class="text-gray-600">admin / admin123</p>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../utils/api'

export default {
  name: 'Login',
  data() {
    return {
      showRegister: false,
      loading: false,
      error: null,
      success: null,
      loginForm: {
        username: '',
        password: ''
      },
      registerForm: {
        username: '',
        email: '',
        password: ''
      }
    }
  },
  methods: {
    async handleLogin() {
      this.loading = true
      this.error = null

      try {
        const response = await api.post('/auth/login', this.loginForm)
        localStorage.setItem('authToken', response.data.token)
        localStorage.setItem('user', JSON.stringify(response.data.user))
        this.$router.push('/dashboard')
      } catch (error) {
        this.error = error.response?.data?.message || 'Login failed'
      } finally {
        this.loading = false
      }
    },

    async handleRegister() {
      this.loading = true
      this.error = null
      this.success = null

      try {
        const response = await api.post('/auth/register', this.registerForm)
        this.success = 'Registration successful! Logging you in...'

        setTimeout(() => {
          localStorage.setItem('authToken', response.data.token)
          localStorage.setItem('user', JSON.stringify(response.data.user))
          this.$router.push('/dashboard')
        }, 1500)
      } catch (error) {
        this.error = error.response?.data?.message || 'Registration failed'
      } finally {
        this.loading = false
      }
    }
  }
}
</script>
