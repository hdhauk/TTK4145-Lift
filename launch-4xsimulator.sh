echo "Launching 4 simulators"
cd hw/simulators/simulator1-53566
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+100
cd ../simulator2-53567
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+300
cd ../simulator3-53568
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+500
cd ../simulator4-53569
gnome-terminal -e "rdmd sim_server.d" --geometry 50x9+200+700
cd ../../..

echo "Waiting..."
sleep .5
echo "..for.."
sleep .5
echo "..the.."
sleep .5
echo "..simulators.."
sleep .5
echo ".. to boot!"

gnome-terminal -e './ttk4145-elevator -sim 53566' --geometry 90x9+680+100
gnome-terminal -e './ttk4145-elevator -sim 53567' --geometry 90x9+680+300
gnome-terminal -e './ttk4145-elevator -sim 53568' --geometry 90x9+680+500
gnome-terminal -e './ttk4145-elevator -sim 53569' --geometry 90x9+680+700
