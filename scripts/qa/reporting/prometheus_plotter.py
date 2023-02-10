# pip install numpy pandas matplotlib requests 

import sys
import os

import matplotlib as mpl
import matplotlib.pyplot as plt

import numpy as np 
import pandas as pd 

import requests 
from urllib.parse import urljoin

from prometheus_pandas import query

release = 'v0.34.27'
path = os.path.join('imgs','cmt2tm1')
prometheus = query.Prometheus('http://localhost:9090')

# Do prometheus queries
queries = [ 
    (( 'cometbft_p2p_peers',                        '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'peers',            dict(ylabel='Peers',     xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_mempool_size',                     '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'mempool_size',     dict(ylabel='TXs',       xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True, kind='area',stacked=True)),
    (( 'avg(cometbft_mempool_size)',                '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'avg_mempool_size', dict(ylabel='TXs',       xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_rounds',                 '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'rounds',           dict(ylabel='# Rounds',  xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_height',                 '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'blocks',           dict(ylabel='# Blocks',  xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)), 
    (( 'rate(cometbft_consensus_height[1m])*60',    '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'block_rate',       dict(ylabel='Blocks/s',  xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_total_txs',              '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'total_txs',        dict(ylabel='# TXs',     xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'rate(cometbft_consensus_total_txs[1m])*60', '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'total_txs_rate',   dict(ylabel='TXs/s',     xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'process_resident_memory_bytes',             '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'memory',           dict(ylabel='Bytes',     xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'avg(process_resident_memory_bytes)',        '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'avg_memory',       dict(ylabel='Bytes',     xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
    (( 'node_load1',                                '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'cpu',              dict(ylabel='CPU Time',  xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)), 
    (( 'avg(node_load1)',                           '2023-02-08T13:09:20Z', '2023-02-08T13:14:20Z', '1s'), 'avg_cpu',          dict(ylabel='CPU Time',  xlabel='time (s)', title='Peers', legend=False, figsize=(10,6), grid=True)),
]

for (query, file_name, pandas_params)  in queries:
    print(query)

    data_frame = prometheus.query_range(*query)
    #Tweak the x ticks
    data_frame = data_frame.set_index(pd.to_timedelta(data_frame.index.strftime('%H:%M:%S')))

    data_frame.plot(**pandas_params)

    plt.savefig(os.path.join(path, file_name + '.png'))
    plt.show()

