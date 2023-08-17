import sys
import networkx as nx
import matplotlib.pyplot as plt
import topology

def plot_topology(peers: dict[int, list[int]], title=""):
    # Create a directed graphs.
    G = nx.DiGraph()

    # Connect the graph.
    color_map = []
    for id, ps in peers.items():
        G.add_node(id)
        for peer_id in ps:
            G.add_edge(id, peer_id)

    # Create the list of colors for each node.
    for gnode in G:
        for id, ps in peers.items():
            if id == gnode:
                color_map.append('lightblue')

    # A dictionary with nodes as keys and positions as values.
    # It fixes the node positions using a fixed seed.
    pos = nx.spring_layout(G, seed=42)  
    
    # Create a figure, implicitly used by other methods.
    plt.figure(figsize=(8, 6))
    
    options = {
        "with_labels": True, 
        "node_size": 500, 
        "node_color": color_map, 
        "font_size": 10, 
        "font_color": 'black',
        "arrows": False,
    }
    
    # Draw the graph using Matplotlib on the existing figure.
    nx.draw_networkx(G, pos, **options)

    plt.title(title)
    plt.axis('off')
    plt.show()


def plot_from_file(path):
    peers = topology.from_file(path)
    plot_topology(peers, title = f"num_nodes={len(peers)}")


if __name__ == "__main__":
    # parse arguments
    topology_json_path = str(sys.argv[1]) if len(sys.argv) > 1 else ""
    
    plot_from_file(topology_json_path)
