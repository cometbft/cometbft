# pip install numpy pandas matplotlib requests 

import sys
import os

import matplotlib as mpl
import matplotlib.pyplot as plt
import matplotlib.dates as md

import numpy as np 
import pandas as pd 

import requests 
from urllib.parse import urljoin

from prometheus_pandas import query

#release = 'v0.37.0-alpha.2'
release = 'v0.38.0-alpha.2'
path = os.path.join('imgs')
prometheus = query.Prometheus('http://localhost:9090')

# Time window
#window_size = dict(seconds=150) #CMT 0.37.x-alpha3
#window_size = dict(seconds=126) #TM v0.37 (200 nodes) baseline
#window_size = dict(hours=1, minutes=28, seconds=25) #TM v0.37.0-alpha.2 (rotating)
#window_size = dict(seconds=130) #homogeneous
#window_size = dict(seconds=127) #baseline
#window_size = dict(seconds=115) #CMT v0.38.0-alpha.2 (200 nodes)
#window_size = dict(hours=1, minutes=46) #CMT v0.38.0-alpha.2 (rotating)
window_size = dict(seconds=150) #CMT v0.38.0-alpha.2 (ve baseline)

ext_window_size = dict(seconds=200)

# Use the time provided by latency_plotter for the selected experiment.
#left_end = '2023-02-08T13:12:20Z' #cmt2 tm1
#left_end = '2023-02-08T10:31:50Z' #cmt1 tm2
#left_end = '2023-02-14T15:18:00Z' #cmt1 tm1
#left_end = '2023-02-07T18:07:00Z' #homogeneous
#left_end = '2022-10-13T19:41:23Z' #baseline
#left_end = '2023-02-22T18:56:29Z' #CMT v0.37.x-alpha3
#left_end = '2022-10-13T15:57:50Z' #TM v0.37 (200 nodes) baseline
#left_end = '2023-03-20T19:45:35Z' #feature/abci++vef merged with main (7d8c9d426)
#left_end = '2023-05-22T09:39:20Z' #CMT v0.38.0-alpha.2 - 200 nodes
#left_end = '2022-10-10T15:47:15Z' #TM v0.37.0-alpha.2 - rotating
#left_end = '2023-05-23T08:09:50Z' #CMT v0.38.0-alpha.2 - rotating

#left_end = '2023-05-25T18:18:04Z' #CMT v0.38.0-alpha.2 - ve baseline
#left_end = '2023-05-30T19:05:32Z' #CMT v0.38.0-alpha.2 - ve 2k
left_end = '2023-05-30T20:44:46Z' #CMT v0.38.0-alpha.2 - ve 4k
#left_end = '2023-05-25T19:42:08Z' #CMT v0.38.0-alpha.2 - ve 8k
#left_end = '2023-05-26T00:28:12Z' #CMT v0.38.0-alpha.2 - ve 16k
#left_end = '2023-05-26T02:12:27Z' #CMT v0.38.0-alpha.2 - ve 32k

useManualrightEnd = False
if useManualrightEnd: 
   #right_end = '2023-05-25T18:54:04Z' #CMT v0.38.0-alpha.2 - ve baseline
   #right_end = '2023-05-30T19:40:41Z' #CMT v0.38.0-alpha.2 - ve 2k
   right_end = '2023-05-30T21:15:37Z' #CMT v0.38.0-alpha.2 - ve 4k
   #right_end = '2023-05-25T20:16:00Z' #CMT v0.38.0-alpha.2 - ve 8k
   #right_end = '2023-05-26T01:01:57Z' #CMT v0.38.0-alpha.2 - ve 16k 
   #right_end = '2023-05-26T02:46:19Z' #CMT v0.38.0-alpha.2 - ve 32k 
   time_window = (left_end, right_end)
else:
   right_end = pd.to_datetime(left_end) + pd.Timedelta(**window_size)
   time_window = (left_end, right_end.strftime('%Y-%m-%dT%H:%M:%SZ'))

ext_right_end = pd.to_datetime(left_end) + pd.Timedelta(**ext_window_size)
ext_time_window = (left_end, ext_right_end.strftime('%Y-%m-%dT%H:%M:%SZ'))


fork='cometbft'
#fork='tendermint'

# Do prometheus queries, depending on the test case
queries200Nodes = [
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

queriesRotating = [
    (( 'rate(' + fork + '_consensus_height[20s])*60',    time_window[0], time_window[1], '1s'), 'rotating_block_rate',    dict(ylabel='blocks/min',     xlabel='time', title='Rate of Block Creation',         legend=False, figsize=(10,6), grid=True), False),
    (( 'rate(' + fork + '_consensus_total_txs[20s])*60', time_window[0], time_window[1], '1s'), 'rotating_txs_rate',      dict(ylabel='TXs/min',        xlabel='time', title='Rate of Transaction processing', legend=False, figsize=(10,6), grid=True), False),
    (( fork + '_consensus_height{job=~"ephemeral.*"} or ' + fork + '_blocksync_latest_block_height{job=~"ephemeral.*"}',
                                                         time_window[0], time_window[1], '1s'), 'rotating_eph_heights',   dict(ylabel='height',         xlabel='time', title='Heights of Ephemeral Nodes',     legend=False, figsize=(10,6), grid=True), False),
    (( fork + '_p2p_peers',                              time_window[0], time_window[1], '1s'), 'rotating_peers',         dict(ylabel='# peers',        xlabel='time', title='Peers',                          legend=False, figsize=(10,6), grid=True), False),
    (( 'avg(process_resident_memory_bytes)',             time_window[0], time_window[1], '1s'), 'rotating_avg_memory',    dict(ylabel='memory (bytes)', xlabel='time', title='Average Memory Usage',           legend=False, figsize=(10,6), grid=True), False),
    (( 'node_load1',                                     time_window[0], time_window[1], '1s'), 'rotating_cpu',           dict(ylabel='load',           xlabel='time', title='Node Load',                      legend=False, figsize=(10,6), grid=True), False),
]

queriesVExtension= [
    (( fork + '_mempool_size',                     time_window[0], time_window[1], '1s'), 'mempool_size',              dict(ylabel='TXs',               xlabel='time (s)', title='Mempool Size',                   legend=False, figsize=(10,6), grid=True, kind='area',stacked=True), False),
    (( fork + '_mempool_size',                     time_window[0], time_window[1], '1s'), 'mempool_size_not_stacked',              dict(ylabel='TXs',               xlabel='time (s)', title='Mempool Size',                   legend=False, figsize=(10,6), grid=True, stacked=False), False),
    (( fork + '_p2p_peers',                        time_window[0], time_window[1], '1s'), 'peers',                     dict(ylabel='# Peers',           xlabel='time (s)', title='Peers',                          legend=False, figsize=(10,6), grid=True), True),
    (( 'avg(' + fork + '_mempool_size)',                time_window[0], time_window[1], '1s'), 'avg_mempool_size',          dict(ylabel='TXs',               xlabel='time (s)', title='Average Mempool Size',           legend=False, figsize=(10,6), grid=True), False),
    (( fork + '_consensus_rounds',                 time_window[0], time_window[1], '1s'), 'rounds',                    dict(ylabel='# Rounds',          xlabel='time (s)', title='Rounds per block',               legend=False, figsize=(10,6), grid=True), False),
    (( 'process_resident_memory_bytes',             time_window[0], time_window[1], '1s'), 'memory',                    dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Memory usage',                   legend=False, figsize=(10,6), grid=True), False),
    (( 'avg(process_resident_memory_bytes)',        time_window[0], time_window[1], '1s'), 'avg_memory',                dict(ylabel='Memory (bytes)',    xlabel='time (s)', title='Average Memory usage',           legend=False, figsize=(10,6), grid=True), False),
    (( 'node_load1',                                time_window[0], time_window[1], '1s'), 'cpu',                       dict(ylabel='Load',              xlabel='time (s)', title='Node load',                      legend=False, figsize=(10,6), grid=True), False), 
    (( 'avg(node_load1)',                           time_window[0], time_window[1], '1s'), 'avg_cpu',                   dict(ylabel='Load',              xlabel='time (s)', title='Average Node load',              legend=False, figsize=(10,6), grid=True), False),
    (( fork + '_consensus_height',                 time_window[0], time_window[1], '1s'), 'blocks',            dict(ylabel='# Blocks',          xlabel='time (s)', title='Blocks in time',                 legend=False, figsize=(10,6), grid=True), False), 
    (( 'rate(' + fork + '_consensus_height[20s])*60',    time_window[0], time_window[1], '1s'), 'block_rate',        dict(ylabel='Blocks/min',        xlabel='time (s)', title='Rate of block creation',         legend=False, figsize=(10,6), grid=True), True),
    (( fork + '_consensus_total_txs',              time_window[0], time_window[1], '1s'), 'total_txs',         dict(ylabel='# TXs',             xlabel='time (s)', title='Transactions in time',           legend=False, figsize=(10,6), grid=True), False),
    (( 'rate(' + fork + '_consensus_total_txs[20s])*60', time_window[0], time_window[1], '1s'), 'total_txs_rate',    dict(ylabel='TXs/min',           xlabel='time (s)', title='Rate of transaction processing', legend=False, figsize=(10,6), grid=True), True),
]

#queries = queries200Nodes
#queries = queriesRotating
queries = queriesVExtension


for (query, file_name, pandas_params, plot_average) in queries:
    print(query)

    data_frame = prometheus.query_range(*query)
    #Tweak the x ticks
    data_frame = data_frame.set_index(md.date2num(data_frame.index))


    pandas_params["title"] += "  -  " + release
    ax = data_frame.plot(**pandas_params)
    if plot_average:
        average = data_frame.mean(axis=1)
        data_frame['__average__'] = average
        pandas_params['lw'] = 8
        pandas_params['style'] = ['--']
        pandas_params['color'] = ['red']
        ax = data_frame['__average__'].plot(**pandas_params)

    ax.xaxis.set_major_formatter(md.DateFormatter('%H:%M:%S'))
    plt.savefig(os.path.join(path, file_name + '.png'))
    plt.plot()

plt.show()
