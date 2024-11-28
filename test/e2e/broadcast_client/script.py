import sys
import pandas as pd

txs = {}

with open(sys.argv[1]) as file:
    for line in file:
        line = line.split(';')
        if line[0] == 'NewBlock':
            ts = line[5].strip().split(',')[:-1]
            if ts != []:
                for t in ts:
                    if t not in txs:
                        txs[t] = {}
                    if 'blocktime' not in txs[t]:
                        txs[t]['blocktime'] = line[3]
                        txs[t]['blockrcv'] = line[4]
        elif line[0] == 'transaction':
            t = line[4].strip()
            if t not in txs:
                txs[t] = {}
            txs[t]['response'] = line[5].strip()

blocktime = []
blockrcv = []
response = []

for tx in txs:
    start = pd.to_datetime(tx)
    try:
        b = pd.to_datetime(txs[tx]['blocktime']) -start
        br = pd.to_datetime(txs[tx]['blockrcv']) -start
        r = pd.to_datetime(txs[tx]['response']) -start
    except:
        continue
    blocktime += [b.microseconds/1000]
    blockrcv += [br.microseconds/1000]
    response += [r.microseconds/1000]

def avrg(l):
    return sum(l)/len(l)
    
print(f'\t{len(blocktime)} transactions')
print(f'Blocktime delta: min {min(blocktime)}ms avrg {avrg(blocktime)}ms max {max(blocktime)}ms')
print(f'Block latency delta: min {min(blockrcv)}ms avrg {avrg(blockrcv)}ms max {max(blockrcv)}ms')
print(f'Send response delta: min {min(response)}ms avrg {avrg(response)}ms max {max(response)}ms')
