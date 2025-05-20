#!/bin/bash

rm -f /tmp/in_* /tmp/out_*

mkfifo /tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
mkfifo /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2
mkfifo /tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3

pids=()

cleanup() {
  echo "Arrêt des processus..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null
  done
  rm -f /tmp/in_* /tmp/out_*
  exit
}

# CTRL+C ou fermeture entraine le cleanup
trap cleanup SIGINT SIGTERM EXIT

# lancement des processus + stockage des PIDs



go run app/*.go -n "app_1"  < /tmp/in_A1 > /tmp/out_A1 & pids+=($!)
go run ctrl/ctrl.go -n "ctrl_1" < /tmp/in_C1 > /tmp/out_C1 & pids+=($!)

go run app/*.go -n "app_2"  < /tmp/in_A2 > /tmp/out_A2 & pids+=($!)
go run ctrl/ctrl.go -n "ctrl_2" < /tmp/in_C2 > /tmp/out_C2 & pids+=($!)

go run app/*.go -n "app_3"  < /tmp/in_A3 > /tmp/out_A3 & pids+=($!)
go run ctrl/ctrl.go -n "ctrl_3" < /tmp/in_C3 > /tmp/out_C3 & pids+=($!)

cd front || exit

npm run dev:8080 & pids+=($!)
npm run dev:8081 & pids+=($!)
npm run dev:8082 & pids+=($!)

#firefox http://localhost:5173
#firefox http://localhost:5174
#firefox http://localhost:5175

cd ..


# Connexions des flux
cat /tmp/out_A1 > /tmp/in_C1 & pids+=($!)
cat /tmp/out_C1 | tee /tmp/in_A1 > /tmp/in_C2 & pids+=($!)

cat /tmp/out_A2 > /tmp/in_C2 & pids+=($!)
cat /tmp/out_C2 | tee /tmp/in_A2 > /tmp/in_C3 & pids+=($!)

cat /tmp/out_A3 > /tmp/in_C3 & pids+=($!)
cat /tmp/out_C3 | tee /tmp/in_A3 > /tmp/in_C1 & pids+=($!)


wait
