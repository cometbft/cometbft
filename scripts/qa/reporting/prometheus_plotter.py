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

release = 'v0.34.x-baseline'
path = os.path.join('imgs')
prometheus = query.Prometheus('http://localhost:9090')

# Time window
window_size = dict(seconds=150)
#window_size = dict(seconds=130) #homogeneous
#window_size = dict(seconds=1270) #baseline
ext_window_size = dict(seconds=180)

#right_end = '2023-02-08T13:14:50Z' #cmt2 tm1
#right_end = '2023-02-08T10:33:50Z' #cmt1 tm2
right_end = '2023-02-14T15:20:30.000Z' #cmt1 tm1
#right_end = '2023-02-07T18:09:10Z' #homogeneous
#right_end = '2022-10-13T19:43:30Z' #baseline
left_end = pd.to_datetime(right_end) - pd.Timedelta(**window_size)
time_window = (left_end.strftime('%Y-%m-%dT%H:%M:%SZ'), right_end)

ext_left_end = pd.to_datetime(right_end) - pd.Timedelta(**ext_window_size)
ext_time_window = (ext_left_end.strftime('%Y-%m-%dT%H:%M:%SZ'), right_end)

fork='cometbft'
#fork='tendermint'

# Do prometheus queries
queries = [ 
    (( fork + '_mempool_size',                     time_window[0], time_window[1], '1s'), 'mempool_size',              dict(ylabel='TXs',               xlabel='time (s)', title='Mempool Size',                   legend=False, figsize=(10,6), grid=True, kind='area',stacked=True), False),
    (( fork + '_p2p_peers',                        time_window[0], time_window[1], '1s'), 'peers',                     dict(ylabel='# Peers',           xlabel='time (s)', title='Peers',                          legend=False, figsize=(10,6), grid=True), True),
    (( 'avg(' + fork + '_mempool_size)',                time_window[0], time_window[1], '1s'), 'avg_mempool_size',          dict(ylabel='TXs',               xlabel='time (s)', title='Average Mempool Size',           legend=False, figsize=(10,6), grid=True), False),
    #(( 'cometbft_consensus_height',                 time_window[0], time_window[1], '1s'), 'blocks_regular',           dict(ylabel='# Blocks',          xlabel='time (s)', title='Blocks in time',                 legend=False, figsize=(10,6), grid=True), False), 
    (( fork + '_consensus_rounds',                 time_window[0], time_window[1], '1s'), 'rounds',                    dict(ylabel='# Rounds',          xlabel='time (s)', title='Rounds per block',               legend=False, figsize=(10,6), grid=True), False),
    (( 'rate(' + fork + '_consensus_height[20s])*60',    time_window[0], time_window[1], '1s'), 'block_rate_regular',        dict(ylabel='Blocks/min',        xlabel='time (s)', title='Rate of block creation',         legend=False, figsize=(10,6), grid=True), True),
    #(( 'avg(rate(cometbft_consensus_height[20s])*60)',    time_window[0], time_window[1], '1s'), 'block_rate_avg_reg',   dict(ylabel='Blocks/min',        xlabel='time (s)', title='Rate of block creation',         legend=False, figsize=(10,6), grid=True), False),
    #(( 'cometbft_consensus_total_txs',              time_window[0], time_window[1], '1s'), 'total_txs_regular',        dict(ylabel='# TXs',             xlabel='time (s)', title='Transactions in time',           legend=False, figsize=(10,6), grid=True), False),
    (( 'rate(' + fork + '_consensus_total_txs[20s])*60', time_window[0], time_window[1], '1s'), 'total_txs_rate_regular',    dict(ylabel='TXs/min',           xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True), True),
    #(( 'avg(rate(cometbft_consensus_total_txs[20s])*60)', time_window[0], time_window[1], '1s'), 'total_txs_rate_avg_reg',   dict(ylabel='TXs/min',       xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True), False),
    (( 'process_resident_memory_bytes',             time_window[0], time_window[1], '1s'), 'memory',                    dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Memory usage',                   legend=False, figsize=(10,6), grid=True), False),
    (( 'avg(process_resident_memory_bytes)',        time_window[0], time_window[1], '1s'), 'avg_memory',                dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Average Memory usage',           legend=False, figsize=(10,6), grid=True), False),
    (( 'node_load1',                                time_window[0], time_window[1], '1s'), 'cpu',                       dict(ylabel='Load',              xlabel='time (s)', title='Node load',                      legend=False, figsize=(10,6), grid=True), False), 
    (( 'avg(node_load1)',                           time_window[0], time_window[1], '1s'), 'avg_cpu',                   dict(ylabel='Load',              xlabel='time (s)', title='Average Node load',              legend=False, figsize=(10,6), grid=True), False),
    #extended window metrics
    (( fork + '_consensus_height',                 ext_time_window[0], ext_time_window[1], '1s'), 'blocks',            dict(ylabel='# Blocks',          xlabel='time (s)', title='Blocks in time',                 legend=False, figsize=(10,6), grid=True), False), 
    (( 'rate(' + fork + '_consensus_height[20s])*60',    ext_time_window[0], ext_time_window[1], '1s'), 'block_rate',        dict(ylabel='Blocks/min',        xlabel='time (s)', title='Rate of block creation',         legend=False, figsize=(10,6), grid=True), True),
    (( fork + '_consensus_total_txs',              ext_time_window[0], ext_time_window[1], '1s'), 'total_txs',         dict(ylabel='# TXs',             xlabel='time (s)', title='Transactions in time',           legend=False, figsize=(10,6), grid=True), False),
    (( 'rate(' + fork + '_consensus_total_txs[20s])*60', ext_time_window[0], ext_time_window[1], '1s'), 'total_txs_rate',    dict(ylabel='TXs/min',           xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True), True),
]

for (query, file_name, pandas_params, plot_average)  in queries:
    print(query)

    data_frame = prometheus.query_range(*query)
    #Tweak the x ticks
    delta_index = pd.to_timedelta(data_frame.index.strftime('%H:%M:%S'))
    data_frame = data_frame.set_index(delta_index)

    data_frame.plot(**pandas_params)
    if plot_average:
        average = data_frame.mean(axis=1)
        data_frame['__average__'] = average
        pandas_params['lw'] = 8
        pandas_params['style'] = ['--']
        pandas_params['color'] = ['red']
        data_frame['__average__'].plot(**pandas_params)

    plt.savefig(os.path.join(path, file_name + '.png'))
    plt.plot()

plt.show()
