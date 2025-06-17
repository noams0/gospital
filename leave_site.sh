#!/bin/bash

# Usage : ./leave_site.sh <pid_app> <pid_ctrl> <pid_net> <site_id>
pid_app=$1
pid_ctrl=$2
pid_net=$3
site_id=$4

if [ -z "$pid_app" ] || [ -z "$pid_ctrl" ] || [ -z "$pid_net" ] || [ -z "$site_id" ]; then
  echo "Usage : $0 <pid_app> <pid_ctrl> <pid_net> <site_id>"
  exit 3
fi

echo "Suppression du site $site_id..."

# Tuer les processus Go
kill "$pid_app" 2>/dev/null
kill "$pid_ctrl" 2>/dev/null
kill "$pid_net" 2>/dev/null

# Tuer les processus cat/tee associés
#[ -f /tmp/pidA$site_id ] && kill $(cat /tmp/pidA$site_id) 2>/dev/null && rm /tmp/pidA$site_id
#[ -f /tmp/pidC$site_id ] && kill $(cat /tmp/pidC$site_id) 2>/dev/null && rm /tmp/pidC$site_id
#[ -f /tmp/pidN$site_id ] && kill $(cat /tmp/pidN$site_id) 2>/dev/null && rm /tmp/pidN$site_id
#
## Supprimer les FIFOs
#rm -f /tmp/in_A$site_id /tmp/out_A$site_id
#rm -f /tmp/in_C$site_id /tmp/out_C$site_id
#rm -f /tmp/in_N$site_id /tmp/out_N$site_id

# Supprimer l'entrée dans les successeurs si elle existe
#[ -f /tmp/succ/$site_id ] && rm /tmp/succ/$site_id

echo "Site $site_id supprimé proprement."
