import random
import sys

import graph
from plot import plot_topology


def other_nodes(id, ids, peers, max):
    """
    Other nodes from `ids` not connected to `id`, who still have connections available.
    """
    return [i for i in ids if id != i and i not in peers[id] and len(peers[i]) < max] 


def connect(id, peers, ids, n):
    num_connections = 0

    # check if the node still has connections available
    num_existing_peers = len(peers[id])
    num_peers_to_add = n - num_existing_peers
    if num_peers_to_add <= 0:
        return peers, 0

    # other nodes not connected to this node, with connections available
    other_nodes_id = other_nodes(id, ids, peers, n)
    if not other_nodes_id:
        return peers, 0
    
    # NOTE: this node may end up with less than n connections, because the other
    # nodes have already reached the max number of connections available for
    # them.
    if num_peers_to_add > len(other_nodes_id):
        num_peers_to_add = len(other_nodes_id)
    
    # make new connections to random nodes
    for peer_id in random.sample(other_nodes_id, num_peers_to_add):
        peers[id].add(peer_id)
        peers[peer_id].add(id)
        num_connections += 1    

    return peers, num_connections


def gen(num_nodes: int, min_peers_per_node: int, max_peers_per_node: int) -> tuple[dict[int, list[int]], int, str]:
    """
    Return a dict from node id to the ids of its peers, and the number of
    connections generated.
    """
    if max_peers_per_node < min_peers_per_node:
        return dict(), 0, ""

    ids = list(range(1, num_nodes + 1))

    # initialize result
    peers: dict[int, set[int]] = dict()
    for id in ids:
        peers[id] = set()

    # initialize counter
    num_connections = 0
    
    # for each node make at most min_peers_per_node connections
    for id in ids:
        peers, num_cons = connect(id, peers, ids, min_peers_per_node)
        num_connections += num_cons

    # try to connect up to max_peers_per_node nodes
    for id in ids:
        up_to = random.randint(min_peers_per_node, max_peers_per_node)
        peers, num_cons = connect(id, peers, ids, up_to)
        # print(f"peers connections': {[len(v) for _,v in peers.items()]}")
        num_connections += num_cons
    
    # convert sets to sorted lists
    peers_list = [(n, sorted(list(ps))) for n, ps in peers.items()]
    
    # if it's not fully connected, try again
    if not graph.is_connected(dict(peers_list)):
        return gen(num_nodes, min_peers_per_node, max_peers_per_node)
    
    peers_hash = str(abs(hash(repr(peers_list))))[:10]
    
    return dict(peers_list), num_connections, peers_hash


def gen_dump_plot(num_nodes: int, min_peers: int, max_peers: int):
    # generate a random network topology
    peers, num_connections, peers_hash = gen(num_nodes, min_peers, max_peers)
    print(f"# num_connections: {num_connections}")
    print(f"# num connections per node: {[len(v) for _,v in peers.items()]}")
    
    # dump to file
    graph.write_to_json(peers, f"topology_{peers_hash}.json")

    # plot
    title = f"num_nodes={num_nodes}, min_peers={min_peers}, max_peers={max_peers}"
    plot_topology(peers, title)


if __name__ == "__main__":
    # parse arguments
    num_nodes = int(sys.argv[1]) if len(sys.argv) > 1 else 16
    min_peers = int(sys.argv[2]) if len(sys.argv) > 2 else 2
    max_peers = int(sys.argv[3]) if len(sys.argv) > 3 else 5

    gen_dump_plot(num_nodes, min_peers, max_peers)
