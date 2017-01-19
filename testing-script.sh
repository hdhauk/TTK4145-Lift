#!/bin/bash

while :
do
    clear
    cat<<EOF
==============================================================
  Elevator test suite
--------------------------------------------------------------
  Please enter your choice:

  Spawn 4x simulator and elevator controller pairs         (1)
  Send elevator 1 upward                                   (2)
  Option                                                   (3)
  Quit                                                     (Q)
--------------------------------------------------------------
Choose and option:
EOF
    read -n1 -s
    case "$REPLY" in
    "1")
        echo "Spawning simulator and controllers"
        ./launch-4xsimulator.sh
        echo "Select simulators"
        SIM1=$(xdotool selectwindow)
        SIM2=$(xdotool selectwindow)
        SIM3=$(xdotool selectwindow)
        SIM4=$(xdotool selectwindow)
        ;;
    "2")
        echo "Entering UP command to simulator 1"
        xdotool windowactivate --sync $SIM1 key "9"
        ;;
    "3")
        echo "you chose choice 3"
        ;;
    "Q")
        exit
        ;;
    "q")
        echo "case sensitive!!"
        ;;
     * )
        echo "invalid option"
        ;;
    esac
    read -p "Press enter to continue"
done
