# Projet SR05 - programmation d'une application répartie

## Scénario

Post effondrement, une épidémie se propage et touche les hôpitaux d'une région. Les médecins s'affairent d'un hôpital à l'autre pour soigner les malades. Un hôpital peut envoyer un médecin à un autre hôpital.

Donnée partagée entre les sites : **nombre de médecins par hôpital.**

Fonctionnalités principales:
- fonctionnalités de base : construction du réseau et connexion avec le front
- cohérence des réplicats grâce à l’algorithme de file d’attente répartie
- sauvegarde répartie datée grâce à l’algorithme de calcul d’instantanés


## Algorithme de contrôle

On intercale un contrôleur entre chaque application et l'anneau. Le `ctrl` contrôle ainsi l'activité entre l'app et le réseau, il intercepte les messages envoyés et reçus et leur applique un traitement, ici on prend l'exemple de la couleur (un site blanc devient rouge) :
- quand ctrl reçoit un message en provenance de son app `(m)`, il y ajoute des infos de controle `(m, couleur,...)` avant de le transmettre sur l'anneau
- quand ctrl intercepte un message à destination de son app de la forme `(m, couleur...)`, il utilise les infos de contrôle pour mettre à jour les siennes, puis transmet le message  sans le traitement `(m)` à son app

Cet ajout d'un contrôleur permet de s'assurer que le **message** `(m)` n'arrive pas avant le **marqueur** `(couleur,...)` dans le cas d'un réseau non FIFO.


## Cohérence des réplicats

Chaque site connaît le nombre de médecins présents sur les autres sites, c'est notre donnée partagée. Pour assurer la cohérence des réplicats, on utilise l'algorithme de **file d'attente répartie** qui permet à chaque site de gérer son entrée en section critique (SC):
- Quand un site souhaite envoyer un médecin à un autre site, il doit d'abord *demander* et *obtenir* la section critique. Il la *relache* une fois la donnée partagée *mise à jour* (nombre de médecins local décrémenté).
- de même, pour recevoir un médecin, un site doit demander et obtenir la section critique. Il la relache une fois la donnée mise à jour (nombre de médecins local incrémenté).


### Estampilles

Pour la cohérence des réplicats, on utilise les estampilles. Les estampilles K permettent en effet de construire une horloge injective : à chaque action correspond une date unique (H(a_i),i). Les actions peuvent alors être strictement et totalement ordonnées dans une liste; on obtient ainsi une unique observation (ou file d’attente).

Au cours lde l'algorithme, chaque site reçoit tous les messages REQ et LIB de tous les autres sites et construit sa propre file d’attente FIFO grâce aux estampilles. Chaque site prend une décision au regard de sa file d’attente (exclusion mutuelle) : si la requête du site est de type REQ et qu’il a l’estampille la plus ancienne, alors il entre en SC.


### Déroulement de l'algorithme

D'abord, le front déclenche l’envoie d’un message spontané de son back App_i <ins>(début)</ins>: 
- App envoie un message `(SC)` au ctrl et attend
- Ctrl envoie un message de type `(req, horloge locale, n° du site)` sur l'anneau
- quand le ctrl a reçu un message de type `(ack, h, n°)` de chaque site, il informe son app qu'elle a la SC
- App décrémente `medecin-=1` la donnée
- envoie deux messages à son ctrl : `(médecin) à S_j` et `(finSC, réplicat)`

- Ctrl transmet le `(réplicat)` et un message de type `(lib, h, n°)` et envoir le `(médecin) à S_j` sur l'anneau


Si le site n'est pas le destinataire du message `(médecin)`, il le transmet sur le réseau.
Sinon si le site est le destinataire du message, il le traite: 
- Ctrl informe son App de la réception d'un médecin
- l'App demande la section critique à son ctrl et attend `waitingforreceiving()`
- Ctrl envoie un message de type `(req, horloge locale, n° du site)` sur l'anneau
- quand le ctrl a reçu un message de type `(ack, h, n°)` de chaque site, il informe son app qu'elle a la SC
- App incrémente la donnée `medecin+=1`
- App informe son Ctrl quelle relâche la SC et lui transmet son réplicat
- Ctrl transmet le `(réplicat)` et un message de type `(lib, h, n°)` sur l'anneau
- tous les autres sites mettent à jour leur (réplicat) et leur estampille.




## Sauvegarde répartie datée

