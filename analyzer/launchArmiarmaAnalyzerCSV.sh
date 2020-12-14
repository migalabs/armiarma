echo "Launching Armiarma Analyzer form CSV"
source venv/bin/activate
folder="metrics/mainnet-4"
python3 armiarma-analyzer.py "csv" "$folder"/peerstore.json "$folder"/csvs/armiarma-metrics.csv "$folder"/plots 
deactivate
echo "Finish Armiarma Analyzer CSV"