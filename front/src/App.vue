<script setup>
import { onMounted, onUnmounted, ref } from 'vue'

const wsUrl = import.meta.env.VITE_APP_WS_URL
const doctorCounts = ref({})

let socket = null

function demanderSectionCritique() {
  const message = { type: "send" }
  socket.send(JSON.stringify(message))
}

onMounted(() => {
  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    console.log('WebSocket connectée')
  }

  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      doctorCounts.value = data
    } catch (err) {
      console.error('Message non JSON :', event.data)
    }
  }

  socket.onerror = (err) => {
    console.error('Erreur WebSocket:', err)
  }

  socket.onclose = () => {
    console.warn('WebSocket fermée')
  }
})

onUnmounted(() => {
  if (socket) socket.close()
})
</script>

<template>
  <button @click="demanderSectionCritique()">Envoyer un médecin</button>

  <h1>Hôpitaux</h1>

  <div class="hospitals-grid">
    <div class="hospital" v-for="(count, site) in doctorCounts" :key="site">
      <h2>{{ site }}</h2>
      <div class="doctors">
        <span v-for="n in count" :key="n">🧑‍⚕️</span>
      </div>
      <p>{{ count }} médecin(s)</p>
    </div>
  </div>
</template>

<style scoped>
button {
  margin-bottom: 20px;
  padding: 8px 12px;
  font-weight: bold;
}

.hospitals-grid {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
  justify-content: center;
}

.hospital {
  color: black;
  background: #f0f8ff;
  border: 2px solid #0077b6;
  border-radius: 10px;
  padding: 1rem;
  width: 180px;
  text-align: center;
  box-shadow: 2px 2px 10px rgba(0,0,0,0.1);
}

.hospital h2 {
  margin-bottom: 0.5rem;
}

.doctors {
  font-size: 24px;
  margin: 0.5rem 0;
  min-height: 28px;
}
</style>
