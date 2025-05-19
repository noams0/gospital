# Projet SR05 - programmation d'une application répartie

## Scénario

Post effondrement, une épidémie se propage et touche les hôpitaux d'une région. Les médecins s'affairent d'un hôpital à l'autre pour soigner les malades. Un hôpital peut envoyer un médecin à un autre hôpital.

Donnée partagée entre les sites : **nombre de médecins par hôpital.**

Fonctionnalités principales:
- fonctionnalités de base : construction du réseau et connexion avec le front
- cohérence des réplicats grâce à l’algorithme de file d’attente répartie
- sauvegarde répartie datée grâce à l’algorithme de calcul d’instantanés


## Algorithme de contrôle

On intercale un contrôleur entre chaque application et l'anneau. Le `ctrl` contrôle ainsi l'activité entre l'app et le réseau, il intercepte les messages envoyés et reçus et leur applique un traitement, ici on prend l'exemple de la couleur (un site blanc devient rouge).
> **Algorithme de contrôle**
- quand ctrl reçoit un message en provenance de son app `(m)`, il y ajoute des infos de controle `(m, couleur,...)` avant de le transmettre sur l'anneau
- quand ctrl intercepte un message à destination de son app de la forme `(m, couleur...)`, il utilise les infos de contrôle pour mettre à jour les siennes, puis transmet le message  sans le traitement `(m)` à son app

Cet ajout d'un contrôleur permet de s'assurer que le **message** `(m)` n'arrive pas avant le **marqueur** `(couleur,...)` dans le cas d'un réseau non FIFO. On évite ainsi de mettre à mal le processus de diffusion.


## Cohérence des réplicats

Chaque site connaît le nombre de médecins présents sur les autres sites, c'est notre donnée partagée. Pour assurer la cohérence des réplicats, on utilise un algorithme qui permet à chaque site de gérer son entrée en section critique (SC):
> **Algorithme de file d'attente répartie**
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

 La **sauvegarde** consiste à réunir des photos locales de l'état de chaque site. Chaque site capture ainsi son état lors du clic et l’envoi sur le réseau à l'état initiateur de la sauvegarde. 

Le **problème réparti** rencontré est le suivant : les clics n’ont pas lieu en même temps car les sites ne sont pas synchronisés. Comment faire pour construire un état global cohérent? 

> **Algorithme de lestage avec collecte des états locaux**

Cet algorithme permet de diffuser la sauvegarde à partir d'un site initiateur sur tout le réseau et de collecter tous les états locaux capturés sur ce même site initiateur pour contruire alors un état global cohérent. En raison de la complexité de l'algorithme, on décide d'en expliquer la construction par ajout de fonctionnalités.

_______________

### Algorithme de lestage
1. <ins>(Initialisation)</ins> : Les variables de chaque site sont initialisées, en particulier `couleur <- blanc` 

2. Une app reçoit un message de son front lui indiquant de lancer la sauvegarde <ins>(début)</ins>. L'app envoie alors un message `(save)`à son ctrl qui le réceptionne <ins>(réception)</ins>. 

3. Le ctrl met à jour sa couleur <ins>(début)</ins>

4.  Le ctrl applique un traitement au message: il est lesté d'une couleur `(save, rouge)` en vue de diffuser la sauvegarde. Le ctrl le transmettre sur l’anneau <ins>(émission)</ins>.

2. Le premier ctrl de l’anneau intercepte le message <ins>(réception)</ins>, compare sa couleur à celle du message. Comme les couleurs diffèrent : `couleur<-rouge`. Ainsi, lorsqu'un site est prévenu de la sauvegarde, sa couleur passe de Blanc à Rouge. Le ctrl envoie ensuite le message sans marqueur à son app `(save)` et fait circuler sur l'anneau le message `(save, rouge)` pour diffuser la sauvegarde aux sites suivants <ins>(émission)</ins>.

3. le même processus se déroule sur les autres sites jusqu'à ce que le message revienne à l'initiateur, déjà rouge. [FIN]

--------------

### Algorithme de collecte des états locaux
*On ajoute la collecte des états locaux à l'algorithme de lestage, décrit ci-dessus.*

1. <ins>(Initialisation)</ins>. Les variables de chaque site sont initialisées : `initiateur <- false`, EG l'état global `EG_i<- {}` et le nombre d'états attendus `NbEA_i <- 3` 

2. Une app reçoit un message de son front lui indiquant de lancer la sauvegarde <ins>(début)</ins>. L'app envoie alors un message `(save)` <ins>(émission)</ins> à son ctrl qui le réceptionne <ins>(réception)</ins>. 

3. Le ctrl à l'initiative de la sauvegarde met à jour ses variables <ins>(début)</ins>. `initiateur <- True`, EG l'état global `EG_i<- {etatLocal_i}` et le nombre d'états attendus `NbEA_i <- 2` . L'état local contient le nom du site, son horloge vectorielle, ainsi que son nombre de médecins.

4. Le ctrl applique un traitement au message et lui applique un traitement : le message est lesté d'une couleur `(save, rouge)` en vue de diffuser la sauvegarde. Le ctrl le transmettre sur l’anneau <ins>(émission)</ins>.

5. Le premier ctrl de l’anneau intercepte le message <ins>(réception)</ins>, compare sa couleur à celle du message. Comme il est blanc, il met à jour ses variables, dont `EG_i<- {etatLocal_i}` et envoie un message `(état, EG_i)`sur l'anneau à destination du site initiateur <ins>(émission)</ins>. Les autres ctrl font de même.

6. <ins>(réception)</ins>. Le ititiateur réceptionne les uns après les autres les messages `(état, EG_j)`: il ajoute `EG_j`à son ensemble pour former petit à petit l'état global et décrémente `NbEA` jusqu'à que la variable atteigne 0. Tous les états locaux ont alors été reçus. [FIN]



### Gestion des messages prépost

*On ajoute la collecte des messages prépost à l'algorithme de collecte des états locaux, décrit ci-dessus.*

1. <ins>(Initialisation)</ins> : Les variables de chaque site sont initialisées, en particulier `NBMA_i <- 0` 


5. un message prépost est un message envoyé sur l’anneau par un site S_i après que la sauvegarde a été initiée sur un site mais avant que le site S_i ait été prévenu du lancement de la sauvegarde. Ce message en transit sur le canal n’est donc compris dans aucune capture d’état local, il est de couleur blanche. On complète donc l’algorithme pour que ce message soit identifié comme prépost par le premier site rouge sur lequel il arrive. Une fois le message prépost identifié et marqué message.prepost<-true), le site rouge le renvoie sur l’anneau. Chaque site le transfère jusqu'à ce que l'initiateur de la sauvegarde l’intercepte et l’ajoute à l'état global de la sauvegarde. 

Ainsi, si on généralise : etatCanal = ensemble émis(i->j) / ensemble reçu(j<-i)
calculs sur des variables doivent être fait sur chaque site “Si un site sur l'anneau envoie un message après le début de la sauvegarde, mais avant d'être prévenu qu'une sauvegarde a été lancée, on obtient un message prépost. Ce message va être identifié comme prépost par le premier site Jaune sur lequel il va arriver..”

“Il sera ensuite redirigé jusqu'à l'initiateur de la sauvegarde pour être ajouté à l'état global”  
(6) Étant donné que les communications sont FIFO sur l'anneau logique, nous n'avons pas besoin de vérifier que tous les messages préposts sont arrivés. En effet, par définition, un message prépost envoyé par un site le sera toujours avant l'envoi du message état de ce site. Ainsi, l'initiateur peut être sûr qu'il recevra les messages préposts d'un site avant son message état, car aucun message ne peut en doubler un autre (FIFO). Il suffit donc à l'initiateur de compter le nombre d'états qu'il reçoit. Lorsqu'il les a tous reçus, il peut considérer que la sauvegarde est terminée. Cela suppose qu'il connaît à l'avance le nombre de sites présents sur le réseau. Une fois la sauvegarde terminée, il vérifie grâce aux horloges vectorielles que la coupure est cohérente.

(7) A la fin de l’algo, le site initiateur de la sauvegarde a construit un état global du système qui contient 
une liste d'états locaux, il y en a autant qu'il y a de sites sur le réseau.
une liste de messages préposts

PS : si le réseau n'était pas FIFO, on aurait du ajouter une variable `bilan` sur chaque site pour évaluer de façon répartie le nombre de messages prépost.


### Vérification de la cohérence de la sauvegarde
