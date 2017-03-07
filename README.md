
# TTK4145-Lift
A distributed lift controller written in Go, created as a project assignment in TTK4145 Real-Time Programming.

[![Build Status](https://travis-ci.com/hdhauk/TTK4145-Lift.svg?token=y9hAjhVWRxqextVgHFNt&branch=master)](https://travis-ci.com/hdhauk/TTK4145-Lift)
[![Go Report Card](https://goreportcard.com/badge/github.com/hdhauk/TTK4145-Lift)](https://goreportcard.com/report/github.com/hdhauk/TTK4145-Lift)

## Highlights
 - Communication based on the Raft consensus algorithm.
 - Can handle loss of up to half of available nodes without degraded functionality.
 - Support both lift-hardware and simulators
 - `godoc` compliant

## Raft algorithm
For a overview over the Raft consensus algorithm see the excellent visualizations at:

- [The Secret Life of Data](http://thesecretlivesofdata.com/raft/)
- [The Raft Consensus Algorithm](https://raft.github.io/)

## Module documentation
|Module name | Description |
|--- |--|----|
|`driver`[![GoDoc](https://godoc.org/github.com/hdhauk/TTK4145-Lift/driver?status.svg)](https://godoc.org/github.com/hdhauk/TTK4145-Lift/driver)|Package driver provides control of both simulated and actual lifts. The package also provide functionality for handeling internal orderes, as well as taking external orders.||
|`peerdiscovery`[![GoDoc](https://godoc.org/github.com/hdhauk/TTK4145-Lift/peerdiscovery?status.svg)](https://godoc.org/github.com/hdhauk/TTK4145-Lift/peerdiscovery)|Package peerdiscovery provides automatic detection of other peers in the same subnet. It does this by utlizing broadcastmessages over UDP.|
|`globalstate`[![GoDoc](https://godoc.org/github.com/hdhauk/TTK4145-Lift/globalstate?status.svg)](https://godoc.org/github.com/hdhauk/TTK4145-Lift/globalstate)|Package globalstate is wrapper package for Hashicorps' implementation of the Raft consensus protocol. See https://github.com/hashicorp/raft.|
|`statetools`[![GoDoc](https://godoc.org/github.com/hdhauk/TTK4145-Lift/statetools?status.svg)](https://godoc.org/github.com/hdhauk/TTK4145-Lift/statetools)|Package statetools implements costfunctions, and tools necessary to replicate some of the globalstate's functionality offline.

## Installation

### Requirements
* Ubuntu 16.10
* Go 1.5.3
* DMD 2.073.0 (only for using the simulator)

Only tested on these versions, but will likely work unless you have any ancient versions

Not tested on Windows, but the Comedi driver will likely be hard to get working.
The simulator does not work properly in MacOS.

### Prerequisites
* Make sure your Go environment is correctly set up:

  The location of the project should look something like this:
~~~~
  $GOPATH/
    ↳ src/
      ↳ github.com/
        ↳ hdhauk/
          ↳ TTK4145-Lift/
~~~~

### 1 .Install Comedi drivers
Download the drivers from [comedi.org](http://www.comedi.org/download/comedilib-0.10.2.tar.gz).
Extract the tarball and open a terminal in the folder and install the library :
~~~~
./configure
make
sudo make install
~~~~

### 2. Install Go dependencies
The project utilize Hashicorps Raft library.
To download all necessary dependencies open a terminal window in the project folder and run
~~~~
cd $GOPATH/src/github.com/hdhauk/TTK4145-Lift
go get -t ./..
~~~~
or install it directly using
~~~~
go get github.com/hashicorp/raft
~~~~


### 3. Install testing tools (Optional)
~~~~
sudo apt-get update
sudo apt-get install xdotool
~~~~

### 4. Build the project
~~~~
cd $GOPATH/src/github.com/hdhauk/TTK4145-Lift
go build .
~~~~

## Usage

If you are intending to run in simulator mode first start the simulator on the local host:
~~~~
cd $GOPATH/src/github.com/hdhauk/TTK4145-Lift/driver/simulators/simulator1-53566
rdmd sim_server.d
~~~~

Run the project with the command `./TTK4145-Lift`.
The following options are available

|Argument  |Additional variable    | Description|
|------|------------|------------|
|`-nick` | name you want | Option to give the elevator a specific id. If omitted it will use the process id|
|`-sim` | number of the port | When set the controller will start in simulator mode an will attempt to connect to a simulator on the provided port (running on localhost) |
|`-raft`|number of the port used for raft communication| Both the port provided and the one above will be used for communication and needs to be available.|
|`-floors`|number of floors| Used to provide a custom number of floors. Default is 4|


Example: `./TTK4145-Lift -nick MyElevator -sim 53566 -raft 8000 - floors 9`
