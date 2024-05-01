import json
import os.path
import toml

folders = ['build/node1/config', 'build/node2/config', 'build/node3/config']
address_book_file_name = 'addrbook.json'
config_file_name = 'config.toml'
genesis_file_name = 'genesis.json'
malicious_ip = '192.167.10.2'

for folder in folders:
    address_book_file = os.path.join(folder, address_book_file_name)
    with open(address_book_file, 'r') as fle:
        addr_book = json.load(fle)
    with open(address_book_file, 'w') as fle:
        addr_book['addrs'] = [ad for ad in addr_book['addrs'] if ad['addr']['ip'] != malicious_ip]
        json.dump(addr_book, fle, indent=True)

    genesis_file = os.path.join(folder, genesis_file_name)
    with open(genesis_file, 'r') as fle:
        gen_file = json.load(fle)
    with open(genesis_file, 'w') as fle:
        gen_file['consensus_params']['feature']['vote_extensions_enable_height'] = "1"
        json.dump(gen_file, fle, indent=True)

    config_file = os.path.join(folder, config_file_name)
    with open(config_file, 'r') as fle:
        lines = toml.load(fle)
    with open(config_file, 'w') as fle:
        pp = lines['p2p']['persistent_peers']
        pp = ','.join([piece for piece in pp.split(',') if malicious_ip not in piece])
        lines['p2p']['persistent_peers'] = pp
        toml.dump(lines, fle)
