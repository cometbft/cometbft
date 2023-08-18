import sys
import topology
from string import Template

import graph
from plot import plot_topology


def val_name(id):
    return f"validator{str(id):0>3}"


manifest_preamble = """
prometheus = true
load_tx_size_bytes = 1024
load_tx_to_send = 5000
load_tx_batch_size = 10
load_tx_connections = 1
pex = false
log_level = "mempool:debug,*:info"
"""


manifest_node_template = """
[node.$id]
persistent_peers = [$peers]
send_no_load = $send_no_load
mempool_reactor = "cat"
"""


def to_manifest_string(peers_dict, send_only_to: list = ["1"]):
    tmpl = Template(manifest_node_template)
    s = manifest_preamble
    for id, peer_ids in peers_dict.items():
        s += tmpl.substitute(
            id=val_name(id), 
            peers=", ".join([f'"{val_name(p)}"' for p in peer_ids]), 
            send_no_load="false" if id in send_only_to else "true"
        )
    return s


if __name__ == "__main__":
    # parse arguments
    if len(sys.argv) == 2:
        # read the only argument as the path to a json file with the topology
        topology_json_path = str(sys.argv[1])
        peers = graph.from_file(topology_json_path)
        manifest_path = topology_json_path.removesuffix(".json") + ".toml"
        plot_title = f"num_nodes={len(peers)}"

    elif len(sys.argv) == 4:
        # read the arguments as the parameters to generate a topology
        num_nodes = int(sys.argv[1])
        min_peers = int(sys.argv[2])
        max_peers = int(sys.argv[3])

        # generate topology
        peers, num_connections, peers_hash = topology.gen(num_nodes, min_peers, max_peers)
        topology_path = f"topology_{peers_hash}.json"
        graph.write_to_json(peers, topology_path)
        print(f"generated {topology_path}")

        manifest_path = f"testnet_{peers_hash}.toml"
        plot_title = f"num_nodes={num_nodes}, min_peers={min_peers}, max_peers={max_peers}"

    else:
        print("Error parsing arguments")
        exit(1)

    # generate manifest file
    with open(manifest_path, "w") as file:
        file.write(to_manifest_string(peers))
        print(f"generated {manifest_path}")

    # visualize topology
    plot_topology(peers, plot_title)
