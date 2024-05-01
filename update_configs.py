import json
import os.path
import toml


folders = ['build/node1/config', 'build/node2/config', 'build/node3/config']
address_book_file_name = 'addrbook.json'
config_file_name = 'config.toml'
malicious_ip = '192.167.10.2'

for folder in folders:
    address_book_file = os.path.join(folder, address_book_file_name)
    with open(address_book_file, 'r') as fle:
        content = json.load(fle)
    with open(address_book_file, 'w') as fle:
        content['addrs'] = [ad for ad in content['addrs'] if ad['addr']['ip'] != malicious_ip]
        json.dump(content, fle, indent=True)
    config_file = os.path.join(folder, config_file_name)
    with open(config_file, 'r') as fle:
        lines = toml.load(fle)
    with open(config_file, 'w') as fle:
        pp = lines['p2p']['persistent_peers']
        pp = ','.join([piece for piece in pp.split(',') if malicious_ip not in piece])
        lines['p2p']['persistent_peers'] = pp
        toml.dump(lines, fle)

