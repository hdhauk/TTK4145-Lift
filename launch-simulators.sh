echo "Launching 4 simulators"
cd hw/simulators/simulator1-53566
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+100
cd ../simulator2-53567
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+300
cd ../simulator3-53568
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+500
cd ../simulator4-53569
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+700
