echo "Launching 4 simulators"
cd hw/simulators/simulator1-53566
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+100
cd ../simulator2-53567
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+320
cd ../simulator3-53568
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+540
cd ../simulator4-53569
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+760
cd ../../..

echo "Waiting..."
sleep 1
echo "..for.."
sleep 1
echo "..the.."
sleep 1
echo "..simulators.."
sleep 1
echo ".. to boot!"

gnome-terminal -e './ttk4145-elevator -sim 53566 -nick sim53566' --geometry 90x10+680+100
gnome-terminal -e './ttk4145-elevator -sim 53567 -nick sim53567' --geometry 90x10+680+320
gnome-terminal -e './ttk4145-elevator -sim 53568 -nick sim53568' --geometry 90x10+680+540
gnome-terminal -e './ttk4145-elevator -sim 53569 -nick sim53569' --geometry 90x10+680+760
