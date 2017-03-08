/*
Package globalstate is wrapper package for Hashicorps' implementation of the
Raft consensus protocol. See https://github.com/hashicorp/raft.

For a description of how the Raft algorithm works see:
 - http://thesecretlivesofdata.com/raft/
 - https://raft.github.io/
 - https://raft.github.io/raft.pdf

TL;DR:

	Raft provide an algorithm for ensuring consensus in the cluster, which we in
	this project use for keeping track of:
	* Last registered floor for all lifts
	* Whether an lift is at a standstill or moving somewhere
	* What buttons are pressed in each floor.
*/
package globalstate
