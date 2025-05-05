<script setup>
import { onMounted, onUnmounted, ref } from 'vue'

// URL de la WebSocket (tu peux la rendre dynamique si besoin)
const wsUrl = import.meta.env.VITE_APP_WS_URL;
console.log(wsUrl); // Affiche le wsUrl pour confirmation

console.log(wsUrl)

const messages = ref([]) // Stocke les messages reçus

let socket = null

onMounted(() => {
  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    console.log('WebSocket connectée')
  }

  socket.onmessage = (event) => {
    const data = event.data
    console.log('Reçu :', data)
    messages.value.push(data)
  }

  socket.onerror = (err) => {
    console.error('Erreur WebSocket:', err)
  }

  socket.onclose = () => {
    console.warn('WebSocket fermée')
  }
})

onUnmounted(() => {
  if (socket) {
    socket.close()
  }
})
</script>

<template>
  <header>
    <h1>Messages reçus</h1>
  </header>

  <main>
    <ul>
      <li v-for="(msg, index) in messages" :key="index">{{ msg }}</li>
    </ul>
  </main>
</template>
