# E2E Testing-Related Scripts

## Random network topology generator

> Requires Python 3.11 or higher, and the libraries `networkx` and `matplotlib`.

Generate a totally random, fully connected network topology and its associated manifest file:
```
python3 scripts/gen_manifest.py <num_nodes> <min_peers> <max_peers>
```
This will generate a json file with the generated topology and the manifest file.

Generate a manifest from an existing topology file:
```
python3 scripts/gen_manifest.py <topology_json_file>
```

Plot an existing topology file:
```
python3 scripts/plot.py <topology_json_file>
```

