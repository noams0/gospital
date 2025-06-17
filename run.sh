#!/bin/bash

# Nombre total de sites : premier argument ou 3 par défaut
total_sites=${1:-3}

echo "Lancement avec $total_sites sites..."

# Nettoyage des FIFO
rm -f /tmp/in_* /tmp/out_*

# Dossier pour stocker les successeurs
mkdir -p /tmp/succ
rm -f /tmp/succ/*


# Création des FIFO dynamiquement
for i in $(seq 1 $total_sites); do
  mkfifo /tmp/in_A$i /tmp/out_A$i /tmp/in_C$i /tmp/out_C$i /tmp/in_N$i /tmp/out_N$i
done

# Tableau des PIDs à surveiller pour cleanup
pids=()

# Fonction de nettoyage à la fermeture
cleanup() {
  echo "Arrêt des processus..."
  for i in $(seq 1 $total_sites); do
    kill $(cat /tmp/pidA$i) 2>/dev/null
    kill $(cat /tmp/pidC$i) 2>/dev/null
    kill $(cat /tmp/pidN$i) 2>/dev/null
  done
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null
  done
  rm -f /tmp/in_* /tmp/out_*
  exit
}

# Appel automatique de cleanup si interruption
trap cleanup SIGINT SIGTERM EXIT


# Lancement des apps et contrôleurs avec passage du nombre total
for i in $(seq 1 $total_sites); do
  prev=$(( (i - 2 + total_sites) % total_sites + 1 ))
  next=$(( (i % total_sites) + 1 ))

  echo "$next" > /tmp/succ/$i

  route="from=net_$prev:to=ctrl,from=ctrl:to=net_$next"
  go run net/net.go -n "net_$i" --route="$route" < /tmp/in_N$i > /tmp/out_N$i & pids+=($!)
  go run app/*.go -n "app_$i" -total $total_sites < /tmp/in_A$i > /tmp/out_A$i & pids+=($!)
  go run ctrl/*.go -n "ctrl_$i"  -total $total_sites  < /tmp/in_C$i > /tmp/out_C$i & pids+=($!)
done

# Lancement des front-ends
cd front || exit

for i in $(seq 0 $((total_sites - 1))); do
  VITE_SITE_ID=$i npm run dev & pids+=($!)
done

cd ..

# Connexion des flux en anneau unidirectionnel
for i in $(seq 1 $total_sites); do
  next=$(( (i % total_sites) + 1 ))

  cat /tmp/out_A$i > /tmp/in_C$i &
  echo $! > /tmp/pidA$i

  cat /tmp/out_C$i | tee /tmp/in_A$i > /tmp/in_N$i &
  echo $! > /tmp/pidC$i

  cat /tmp/out_N$i | tee /tmp/in_C$i > /tmp/in_N$next &
  echo $! > /tmp/pidN$i
done


wait
