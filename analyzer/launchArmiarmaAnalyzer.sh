echo "Launching Armiarma Analyzer form JSON"
source venv/bin/activate
folder="metrics/mainnet-4"
python3 armiarma-analyzer.py "json" "$folder"/peerstore2.json "$folder"/gossip-metrics2.json "$folder"/plots "$folder"/csvs/armiarma-metrics.csv
deactivate
echo "Finish Armiarma Analyzer JSON"
