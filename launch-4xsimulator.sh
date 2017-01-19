echo "Launching 4 simulators"
cd hw/simulators/simulator1-53566
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+100 --title="sim53566"
cd ../simulator2-53567
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+320 --title="sim53567"
cd ../simulator3-53568
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+540 --title="sim53568"
cd ../simulator4-53569
gnome-terminal -e "rdmd sim_server.d" --geometry 50x10+200+760 --title="sim53569"
cd ../../..

echo "Waiting..."
sleep .2
echo "..for.."
sleep .2
echo "..the.."
sleep .2
echo "..simulators.."
sleep .2
echo ".. to boot!"

gnome-terminal -e './ttk4145-elevator -sim 53566 -nick sim53566' --geometry 90x10+680+100 --title="controller53566"
gnome-terminal -e './ttk4145-elevator -sim 53567 -nick sim53567' --geometry 90x10+680+320 --title="controller53567"
gnome-terminal -e './ttk4145-elevator -sim 53568 -nick sim53568' --geometry 90x10+680+540 --title="controller53568"
gnome-terminal -e './ttk4145-elevator -sim 53569 -nick sim53569' --geometry 90x10+680+760 --title="controller53569"
