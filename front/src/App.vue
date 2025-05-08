<script setup

>function demanderSectionCritique() {
    console.log("demandeSC")
    const message = {
      type: "demandeSC",
    }
  socket.send(JSON.stringify(message))}

import { onMounted, onUnmounted, ref } from 'vue'

const wsUrl = import.meta.env.VITE_APP_WS_URL;
console.log(wsUrl);

console.log(wsUrl)

const messages = ref([])

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
  <button @click="demanderSectionCritique()">Demander SC</button>
  <header>
    <h1>Messages reçus</h1>
  </header>

  <main>
    <ul>
      <li v-for="(msg, index) in messages" :key="index">{{ msg }}</li>
    </ul>
  </main>
</template>
