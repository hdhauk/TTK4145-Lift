# Design Presenstation

## Part 1: Fault tolerance

## System overview
* Master-slave arkitektur
* All heiser vet til envher tid hvilke knapper som er trykket og hvor de andre er, ved hjelp av en global database.
* Alle heiser har også en lokal kø med oversikt over alle tildelte og lokale ordre (knappene inni heisen)

#### Masters oppgaver
* Ta imot oppdateringer om knapper som er trykket, hva heisene driver med og når en bestilling er utført.
* Delegere oppdrag til den best egnede heisen når det er trykket på en knapp ett eller annet sted.
*

#### Slavens oppgaver
* Si ifra hva om hva heisen gjør envher tid.
* Si ifra om hvilke knapper som blir trykket på.
* Utføre tildelte ordrer fra master.
* Dersom man ikke får kontakt med master, utføre ordrer på egenhånd.

### The Raft algorithm
* Algoritme som løser "Distributed consensus"-problemet.
* Sikrer at det til envhver tid blir valgt en korrekt master. Også dersom master går offline.
* Sikrer integriteten til dataen som ligger i den fellese databasen.

### Hva/hvis:
* #### En enhet mister kontakt med nettverket
  * Alle ordrer blir lagt til i lokal kø, etter en timeout hvor når man forsøker å kontakte masteren.
* #### En enhet kræsjer etter å ha blitt tildelt en ordre.
  * Master overvåker tildelte ordrer og hvis en ordre ikke er utført innen rimelig tid, kjøres kostberegninen på nytt (nå uten heisen som er offline) og ordren tildeles som vanlig.
* #### Master kræsjer / og eller eksploderer
 * Raft vil automatisk trigge et nytt ledervalg og den nye lederen til ta over oppgavene til den forrige lederen. Ingen data er mistet siden alle til enhver tid har nøyaktig samme database.

## Part 2: Modules and interfaces
* ### 4 moduler + `main`
 * `peerdiscovery`
 * `globalstate`
 * `driver`
 * `statetools`
