#!/bin/bash

rm -f /tmp/in_* /tmp/out_*

# Création des FIFO pour 5 apps et 5 contrôleurs
for i in {1..5}; do
  mkfifo /tmp/in_A$i /tmp/out_A$i /tmp/in_C$i /tmp/out_C$i
done

pids=()

cleanup() {
  echo "Arrêt des processus..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null
  done
  rm -f /tmp/in_* /tmp/out_*
  exit
}

trap cleanup SIGINT SIGTERM EXIT

# Lancement des apps et des contrôleurs
for i in {1..5}; do
  go run app/*.go -n "app_$i"  < /tmp/in_A$i > /tmp/out_A$i & pids+=($!)
  go run ctrl/*.go -n "ctrl_$i" < /tmp/in_C$i > /tmp/out_C$i & pids+=($!)
done

# Lancement des front-ends
cd front || exit

npm run dev:8080 & pids+=($!)
npm run dev:8081 & pids+=($!)
npm run dev:8082 & pids+=($!)
npm run dev:8083 & pids+=($!)
npm run dev:8084 & pids+=($!)

cd ..

# Connexions des flux (anneau unidirectionnel)
for i in {1..5}; do
  next=$(( (i % 5) + 1 ))

  cat /tmp/out_A$i > /tmp/in_C$i & pids+=($!)
  cat /tmp/out_C$i | tee /tmp/in_A$i > /tmp/in_C$next & pids+=($!)
done

wait
