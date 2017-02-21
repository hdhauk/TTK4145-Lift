# Elevator Cluster
[![CircleCI](https://circleci.com/bb/halvor_haukvik/ttk4145-elevator.svg?style=svg)](https://circleci.com/bb/halvor_haukvik/ttk4145-elevator)

## Installation

### Requirements
* Ubuntu 16.10
* Go 1.5.3
* DMD 2.073.0 (only for using the simulator)

Only tested on these versions, but will likely work unless you have any ancient versions

Not tested on Windows, but the comedi driver will likely be hard to get working.
The simulator does not work properly in MacOS.

### Prerequisites
* Make sure your Go enviroment is correctly set up:

  The location of the project should look something like this:
~~~~
  $GOPATH/
    ↳ src/
      ↳ bitbucket.org/
        ↳ halvor_haukvik/
          ↳ ttk4145-elevator/
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
cd $GOPATH/src/bitbucket.org/halvor_haukvik/ttk4145-elevator
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
cd $GOPATH/src/bitbucket.org/halvor_haukvik/ttk4145-elevator
go build .
~~~~

## Usage

If you are intending to run in simulator mode first start the simulator on the local host:
~~~~
cd $GOPATH/src/bitbucket.org/halvor_haukvik/ttk4145-elevator/driver/simulators/simulator1-53566
rdmd sim_server.d
~~~~

Run the project with the command `./ttk4145-elevator`.
The following options are available

|Argument  |Additional variable    | Description|
|------|------------|------------|
|`-nick` | name you want | Option to give the elevator a specific id. If omitted it will use the process id|
|`-sim` | number of the port | When set the controller will start in simulator mode an will attempt to connect to a simulator on the provided port (running on localhost) |
|`-raft`|number of the port used for raft communication| Both the port provided and the one above will be used for communication and needs to be available.|
|`-floors`|number of floors| Used to provide a custom number of floors. Default is 4|


Example: `./ttk4145-elevator -nick MyElevator -sim 53566 -raft 8000 - floors 9`