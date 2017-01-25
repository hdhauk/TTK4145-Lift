# Elevator Cluster

## Brainstorming general algorithm
* Distributed store
  * Managed by the [Raft algorithm](http://thesecretlivesofdata.com/raft/)
* Consensus about
  * What orders are waiting (ie. what order buttons are currently lit in the hallways)
  * Where all online elevators are
