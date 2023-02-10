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
path = os.path.join('imgs')
prometheus = query.Prometheus('http://localhost:9090')

# Time window
window_size = dict(seconds=120)
ext_window_size = dict(seconds=150)

#right_end = '2023-02-08T13:14:20Z' #cmt2 tm1
right_end = '2023-02-08T10:33:20Z' #cmt1 tm2
left_end = pd.to_datetime(right_end) - pd.Timedelta(**window_size)
time_window = (left_end.strftime('%Y-%m-%dT%H:%M:%SZ'), right_end)

ext_left_end = pd.to_datetime(right_end) - pd.Timedelta(**ext_window_size)
ext_time_window = (ext_left_end.strftime('%Y-%m-%dT%H:%M:%SZ'), right_end)

# Do prometheus queries
queries = [ 
    (( 'cometbft_p2p_peers',                        time_window[0], time_window[1], '1s'), 'peers',            dict(ylabel='Peers',             xlabel='time (s)', title='Peers',                   legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_mempool_size',                     time_window[0], time_window[1], '1s'), 'mempool_size',     dict(ylabel='TXs',               xlabel='time (s)', title='Mempool Size',            legend=False, figsize=(10,6), grid=True, kind='area',stacked=True)),
    (( 'avg(cometbft_mempool_size)',                time_window[0], time_window[1], '1s'), 'avg_mempool_size', dict(ylabel='TXs',               xlabel='time (s)', title='Average Mempool Size',    legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_rounds',                 time_window[0], time_window[1], '1s'), 'rounds',           dict(ylabel='# Rounds',          xlabel='time (s)', title='Rounds per block',        legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_height',                 time_window[0], time_window[1], '1s'), 'blocks_regular',           dict(ylabel='# Blocks',          xlabel='time (s)', title='Blocks in time',          legend=False, figsize=(10,6), grid=True)), 
    (( 'rate(cometbft_consensus_height[1m])*60',    time_window[0], time_window[1], '1s'), 'block_rate_regular',       dict(ylabel='Blocks/s',          xlabel='time (s)', title='Rate of block creation',  legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_total_txs',              time_window[0], time_window[1], '1s'), 'total_txs_regular',        dict(ylabel='# TXs',             xlabel='time (s)', title='Transactions in time',    legend=False, figsize=(10,6), grid=True)),
    (( 'rate(cometbft_consensus_total_txs[1m])*60', time_window[0], time_window[1], '1s'), 'total_txs_rate_regular',   dict(ylabel='TXs/s',             xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True)),
    (( 'process_resident_memory_bytes',             time_window[0], time_window[1], '1s'), 'memory',           dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Memory usage',            legend=False, figsize=(10,6), grid=True)),
    (( 'avg(process_resident_memory_bytes)',        time_window[0], time_window[1], '1s'), 'avg_memory',       dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Average Memory usage',    legend=False, figsize=(10,6), grid=True)),
    (( 'node_load1',                                time_window[0], time_window[1], '1s'), 'cpu',              dict(ylabel='Load',              xlabel='time (s)', title='Node load',               legend=False, figsize=(10,6), grid=True)), 
    (( 'avg(node_load1)',                           time_window[0], time_window[1], '1s'), 'avg_cpu',          dict(ylabel='Load',              xlabel='time (s)', title='Average Node load',       legend=False, figsize=(10,6), grid=True)),
    #extended window metrics
    (( 'cometbft_consensus_height',                 ext_time_window[0], ext_time_window[1], '1s'), 'blocks',           dict(ylabel='# Blocks',          xlabel='time (s)', title='Blocks in time',          legend=False, figsize=(10,6), grid=True)), 
    (( 'rate(cometbft_consensus_height[1m])*60',    ext_time_window[0], ext_time_window[1], '1s'), 'block_rate',       dict(ylabel='Blocks/s',          xlabel='time (s)', title='Rate of block creation',  legend=False, figsize=(10,6), grid=True)),
    (( 'cometbft_consensus_total_txs',              ext_time_window[0], ext_time_window[1], '1s'), 'total_txs',        dict(ylabel='# TXs',             xlabel='time (s)', title='Transactions in time',    legend=False, figsize=(10,6), grid=True)),
    (( 'rate(cometbft_consensus_total_txs[1m])*60', ext_time_window[0], ext_time_window[1], '1s'), 'total_txs_rate',   dict(ylabel='TXs/s',             xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True)),
]

for (query, file_name, pandas_params)  in queries:
    print(query)

    data_frame = prometheus.query_range(*query)
    #Tweak the x ticks
    data_frame = data_frame.set_index(pd.to_timedelta(data_frame.index.strftime('%H:%M:%S')))

    data_frame.plot(**pandas_params)

    plt.savefig(os.path.join(path, file_name + '.png'))
    plt.show()

