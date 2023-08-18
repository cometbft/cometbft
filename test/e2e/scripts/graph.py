import json
import networkx as nx


def to_NxGraph(nodes_edges: dict[int, list[int]]) -> nx.Graph:
    G = nx.Graph()
    for id, peers in nodes_edges.items():
        G.add_node(id)
        for peer_id in peers:
            G.add_edge(id, peer_id)
    return G


def is_connected(nodes_edges: dict[int, list[int]]) -> bool:
    G = to_NxGraph(nodes_edges)
    return nx.is_connected(G)


def write_to_json(nodes_edges: dict[int, list[int]], file_name: str):
    nodes_edges_json = json.dumps(nodes_edges)
    with open(file_name, "w") as file:
        file.write(nodes_edges_json)


def from_file(path) -> dict[int, list[int]]:
    with open(path, "r") as file:
        nodes_edges = dict(json.load(file))
        
        # convert keys from string to int (json keys cannot be integers)
        nodes_edges = dict([(int(k), list(v)) for k,v in nodes_edges.items()])
        
        return nodes_edges