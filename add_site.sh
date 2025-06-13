#!/bin/bash

# Usage : ./add_site.sh <new_id> <attach_to> <total_sites>
new_id=$1
attach_to=$2
total_sites=$3

if [ -z "$new_id" ] || [ -z "$attach_to" ] || [ -z "$total_sites" ]; then
  echo "Usage : $0 <new_site_id> <site_to_attach_after> <new_total_sites>"
  exit 1
fi

pids=()


next_attach=$(cat /tmp/succ/$attach_to)

echo "succ"

echo $next_attach


echo "Ajout du site $new_id entre $attach_to et $next_attach (total = $total_sites)..."

echo "$attach_to" > /tmp/succ/$new_id

# Création des FIFO pour le nouveau site
mkfifo /tmp/in_A$new_id /tmp/out_A$new_id
mkfifo /tmp/in_C$new_id /tmp/out_C$new_id
mkfifo /tmp/in_N$new_id /tmp/out_N$new_id

# Lancement des processus Go pour le nouveau site
go run app/*.go -n "app_$new_id" -total $total_sites < /tmp/in_A$new_id > /tmp/out_A$new_id & pids+=($!)

go run ctrl/*.go -n "ctrl_$new_id" -total $total_sites < /tmp/in_C$new_id > /tmp/out_C$new_id & pids+=($!)


# Kill l'ancien lien net_$attach_to -> net_$next_attach
echo "Reconfiguration de l’anneau entre net_$attach_to -> net_$next_attach"
kill $(cat /tmp/pidN$attach_to)
sleep 0.5


succs=$(cat /tmp/succ/$attach_to)
for s in $succs; do
  outputs+=" /tmp/in_N$s"
done

echo "succs" $outputs
# Reconnecte net_$attach_to -> net_$new_id

cat /tmp/out_N$attach_to | tee /tmp/in_C$attach_to /tmp/in_N$new_id  $outputs > /dev/null &
echo $! > /tmp/pidN$attach_to


# Connecte net_$new_id -> net_$attach_to
cat /tmp/out_N$new_id | tee /tmp/in_C$new_id > /tmp/in_N$attach_to &   echo $! > /tmp/pidN$new_id

# Connexions internes : app -> ctrl -> net
cat /tmp/out_A$new_id > /tmp/in_C$new_id & pids+=($!)

cat /tmp/out_C$new_id | tee /tmp/in_A$new_id > /tmp/in_N$new_id & pids+=($!)

# Lancement du front-end
cd front || exit
VITE_SITE_ID=$((new_id - 1)) npm run dev & pids+=($!)
cd ..


echo "$new_id" >> "/tmp/succ/$attach_to"
echo "Site $new_id ajouté et connecté."

route="from=net_$attach_to:to=ctrl,from=ctrl:to=net_$attach_to"
go run net/net.go -n "net_$new_id" --route="$route" < /tmp/in_N$new_id > /tmp/out_N$new_id  & pids+=($!)

sleep 0.3  # légère attente pour la création effective

cleanup() {
  echo "Interruption : nettoyage..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null
  done
  kill $(cat /tmp/pidN$new_id) 2>/dev/null
  kill $(cat /tmp/pidN$attach_to) 2>/dev/null
#  kill $(cat /tmp/pidC$new_id) 2>/dev/null
#  kill $(cat /tmp/pidA$new_id) 2>/dev/null
#  kill $(cat /tmp/pidN$new_id) 2>/dev/null
#  kill $(cat /tmp/pidN$new_id) 2>/dev/null
#  kill $(cat /tmp/pidN$attach_to) 2>/dev/null
  rm -f /tmp/in_A$new_id /tmp/out_A$new_id
  rm -f /tmp/in_C$new_id /tmp/out_C$new_id
  rm -f /tmp/in_N$new_id /tmp/out_N$new_id
  rm -f /tmp/pidA$new_id /tmp/pidC$new_id
  exit
}

trap cleanup SIGINT SIGTERM EXIT


wait
