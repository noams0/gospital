<script setup>
import {computed, onMounted, onUnmounted, ref} from 'vue'

const doctorCounts = ref({})

let socket = null

const siteId = import.meta.env.VITE_SITE_ID
const wsUrl = getWsUrl(siteId)
console.log('WebSocket URL:', wsUrl)

function getWsUrl(siteId) {
  const port = 8080 + Number(siteId)
  return `ws://localhost:${port}/ws`
}


onMounted(() => {
  socket = new WebSocket(wsUrl)

  socket.onopen = () => {
    console.log('WebSocket connectée')
  }


  socket.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      // console.log(data)
      doctorCounts.value = data.doctors
      doctorCountsSender.value = data.sender
      activityLog.value = data.activity_log || []
      snapshot.value = data.snapshot
      console.log(data)


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
const snapshot = ref("")

const parsedSnapshot = computed(() => {
  try {
    const outer = JSON.parse(snapshot.value || '{}')
    const result = {}

    for (const [site, innerStr] of Object.entries(outer)) {
      // Cas spécial pour PREPOST
      if (site === "PREPOST") {
        result[site] = innerStr
        continue
      }

      try {
        const inner = JSON.parse(innerStr)
        result[site] = inner[site] || inner
      } catch (e) {
        result[site] = `Erreur de parsing: ${innerStr}`
      }
    }

    return result
  } catch (err) {
    return {}
  }
})


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

function demanderSnapshot() {
  const message = {
    type: "snapshot"
  }
  socket.send(JSON.stringify(message))
}

function askToLeave() {
  const message = {
    type: "askToLeave"
  }
  socket.send(JSON.stringify(message))
}

function askToQuit() {
  const message = {
    type: "askToQuit"
  }
  socket.send(JSON.stringify(message))
}


const speed = ref(50)

function changerVitesse() {
  if (socket.readyState === WebSocket.OPEN) {
    const message = JSON.stringify({
      type: 'speed',
      delay: speed.value
    })
    socket.send(message)
  }
}

onUnmounted(() => {
  if (socket) socket.close()
})
</script>

<template>

  <div class="speed-control">
    <label for="speedRange">⏱️ Vitesse de simulation : {{ speed }} ms</label>
    <input
        id="speedRange"
        type="range"
        min="10"
        max="5000"
        step="10"
        v-model="speed"
    />
    <button @click="changerVitesse">✅ Appliquer la vitesse</button>
  </div>

  <button @click="demanderSnapshot" style="margin-bottom: 20px">
    🔄 Déclencher une sauvegarde instantanée
  </button>
  <br>

  <button @click="askToLeave" style="margin-bottom: 20px">
    BYE Quitter le réseau
  </button>

  <button @click="askToQuit" style="margin-bottom: 20px">
    BYE BYE Quitter physiquement le réseau
  </button>
  <div class="hospital-container">
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
  </div>
  <div class="snapshot-display">
    <h3>📸 État global sauvegardé</h3>
    <div class="hospital" v-for="(val, site) in parsedSnapshot" :key="site">
      <h2>{{ site }}</h2>
      <div class="doctors" v-if="site !== 'PREPOST'">
        <span v-for="n in val" :key="n">🧑‍⚕️</span>
      </div>
      <p >{{ val }} médecin(s)</p>
  </div>
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



<!--    <div class="snapshot-site" v-for="(etat, site) in snapshotParsed" :key="site">-->
<!--      <h4>{{ site }}</h4>-->
<!--      <ul>-->
<!--        <li><strong>Horloge :</strong> {{ etat.Horloge }}</li>-->
<!--        <li><strong>InSection :</strong> {{ etat.InSection ? "Oui" : "Non" }}</li>-->
<!--        <li><strong>Doctors :</strong>-->
<!--          <ul>-->
<!--            <li v-for="(count, doc) in etat.DoctorsCount" :key="doc">-->
<!--              {{ doc }} : {{ count }}-->
<!--            </li>-->
<!--          </ul>-->
<!--        </li>-->
<!--      </ul>-->
<!--    </div>-->


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

.hospital-container{
  display: flex;
  ;
}
.hospital {

  color: black;
  background: #f0f8ff;
  border: 2px solid #b60006;
  border-radius: 10px;
  padding: 1rem;
  margin: 5px;
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



.snapshot-display {
  display: flex;
  color: black;
  margin-top: 2rem;
  padding: 1rem;
  background: #e0f7fa;
  border-left: 4px solid #006064;
}

.snapshot-site {
  margin-bottom: 1rem;
  padding: 0.5rem;
  background: #ffffff;
  border: 1px solid #ccc;
  border-radius: 8px;
}

.snapshot-site h4 {
  margin-bottom: 0.3rem;
  color: #00796b;
}

.speed-control {
  margin: 1em 0;
}
input[type="range"] {
  width: 100%;
}

</style>
