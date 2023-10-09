import sys
import networkx as nx
import matplotlib.pyplot as plt
from pyvis.network import Network


import graph


def plot_topology(nodes_edges: dict[int, list[int]], title=""):
    # Create a directed graphs.
    G = graph.to_NxGraph(nodes_edges)

    # Create the list of colors for each node.
    color_map = []
    for gnode in G:
        for id, _ in nodes_edges.items():
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

    nt = Network('1080', '1080')
    nt.from_nx(G)
    nt.toggle_physics(True)
    nt.show_buttons(filter_=['physics'])
    nt.show('nx.html', notebook=False)


def plot_from_file(path):
    nodes_edges = graph.from_file(path)
    plot_topology(nodes_edges, title = f"num_nodes={len(nodes_edges)}")


if __name__ == "__main__":
    # parse arguments
    topology_json_path = str(sys.argv[1]) if len(sys.argv) > 1 else ""
    
    plot_from_file(topology_json_path)
