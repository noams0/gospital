<script setup>
import {computed, onMounted, onUnmounted, ref} from 'vue'

const wsUrl = import.meta.env.VITE_APP_WS_URL
const doctorCounts = ref({})

let socket = null


onMounted(() => {
  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    console.log('WebSocket connectée')
  }


  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      console.log(data)
      doctorCounts.value = data.doctors
      doctorCountsSender.value = data.sender
      activityLog.value = data.activity_log || []
      console.log(activityLog.value)
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
const doctorCountsSender = ref("")
const activityLog = ref("")

const doctorCountsSenderNb = computed(() =>
doctorCounts.value[doctorCountsSender.value]
)

function getLogClass(log) {
  if (log.includes("DemSC")) return "log-dem";
  if (log.includes("DebSC")) return "log-deb";
  if (log.includes("FinSC")) return "log-fin";
  return "log-default";
}

function formatLog(entry) {
  if (entry.startsWith("TAB_REQ")) {
    const rows = entry
        .replace("TAB_REQ", "")
        .split(",")
        .filter(e => e.trim() !== "")
        .map(siteStr => {
          const [site, rest] = siteStr.split(" : ");
          const kvPairs = rest
              ?.split(",")
              .map(p => p.trim())
              .filter(p => p.length > 0);

          const indented = kvPairs?.map(kv => `&nbsp;&nbsp;${kv}`).join("<br>") || "";
          return `<strong>${site.trim()}</strong> <br>${indented}`;
        });

    return `<strong>TAB_REQ</strong><br>${rows.join("<br>")}`;
  } else {
    return entry; // Affichage brut pour les autres logs
  }
}

function envoyerMedecin(site) {
  const message = {
    type: "send",
    to: site
  }
  socket.send(JSON.stringify(message))
}

onUnmounted(() => {
  if (socket) socket.close()
})
</script>

<template>
  <div class="hospital" v-for="(count, site) in doctorCounts" :key="site">
    <h2>{{ site }}</h2>
    <div class="doctors">
      <span v-for="n in count" :key="n">🧑‍⚕️</span>
    </div>
    <p>{{ count }} médecin(s)</p>
    <button
        v-if="site !== doctorCountsSender && doctorCountsSenderNb > 0"
        @click="envoyerMedecin(site)"
    >
      Envoyer un médecin
    </button>
  </div>
  <div class="activity-log">
    <h3>Journal des activités</h3>
    <ul>
      <li
          v-for="(entry, index) in activityLog"
          :key="index"
          :class="getLogClass(entry)"
          v-html="formatLog(entry)"
      />
    </ul>
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
.activity-log {
  margin: 2rem 0;
  padding: 1rem;
  background: #f8f9fa;
  border-left: 4px solid #0077b6;
}

.activity-log h3 {
  margin-bottom: 0.5rem;
  color: black;

}

.activity-log ul {
  list-style: none;
  padding-left: 0;
}

.activity-log li {
  padding: 0.25rem 0.5rem;
  border-radius: 5px;
  margin-bottom: 0.3rem;
  font-weight: bold;
}

.log-dem {
  background-color: #fff3cd;
  color: #856404;
}

.log-deb {
  background-color: #d1ecf1;
  color: #0c5460;
}

.log-fin {
  background-color: #d4edda;
  color: #155724;
}

.log-default {
  background-color: #f8d7da;
  color: #721c24;
}


</style>
