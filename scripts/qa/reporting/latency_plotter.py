import sys
import os
from datetime import datetime
import pytz

import matplotlib as mpl
import matplotlib.pyplot as plt

import numpy as np
import pandas as pd

release = 'v0.38.0-alpha2'

#FIXME: figure out in which timezone prometheus was running to adjust to UTC.
tz = pytz.timezone('America/Sao_Paulo')

if len(sys.argv) != 2:
    print('Pls provide the raw.csv file')
    exit()
else:
    csvpath = sys.argv[1]
    if not os.path.exists(csvpath):
       print('Pls provide a valid the raw.csv file')
       exit()
        
    print(csvpath)

path = os.path.join('imgs')

#Load the CSV
csv = pd.read_csv(csvpath)

#Transform ns to s in the latency/duration
csv['duration_ns'] = csv['duration_ns'].apply(lambda x: x/10**9)
csv['block_time'] = csv['block_time'].apply(lambda x: x/10**9)

#Group by experiment
groups = csv.groupby(['experiment_id'])

#number of rows and columns in the graph
ncols = 2 if groups.ngroups > 1 else 1
nrows = int( np.ceil(groups.ngroups / ncols)) if groups.ngroups > 1 else 1
fig, axes = plt.subplots(nrows=nrows, ncols=ncols, figsize=(6*ncols, 4*nrows), sharey=False)
fig.tight_layout(pad=5.0)


#Plot experiments as subplots 
for (key,ax) in zip(groups.groups.keys(), [axes] if ncols == 1 else axes.flatten()):
    group = groups.get_group(key)
    ax.set_ylabel('latency (s)')
    ax.set_xlabel('experiment time (s)')
    ax.set_title(key)
    ax.grid(True)

    #Group by connection number and transaction rate
    paramGroups = group.groupby(['connections','rate'])
    for (subKey) in paramGroups.groups.keys():
        subGroup = paramGroups.get_group(subKey)
        startTime = subGroup.block_time.min()
        endTime = subGroup.block_time.max()
        localStartTime = tz.localize(datetime.fromtimestamp(startTime)).astimezone(pytz.utc)
        localEndTime  = tz.localize(datetime.fromtimestamp(endTime)).astimezone(pytz.utc)
        subGroup.block_time.apply(lambda x: x - startTime )
        mean = subGroup.duration_ns.mean()
        print('exp', key ,'start', localEndTime.strftime("%Y-%m-%dT%H:%M:%SZ"), 'end', localStartTime.strftime("%Y-%m-%dT%H:%M:%SZ"), 'duration', endTime - startTime, "mean", mean)

        (con,rate) = subKey
        label = 'c='+str(con) + ' r='+ str(rate)
        ax.axhline(y = mean, color = 'r', linestyle = '-', label="mean")
        ax.scatter(subGroup.block_time, subGroup.duration_ns, label=label)
    ax.legend()

    #Save individual axes
    extent = ax.get_window_extent().transformed(fig.dpi_scale_trans.inverted())
    fig.savefig(os.path.join(path,'e_'+key + '.png'), bbox_inches=extent.expanded(1.2, 1.3))

fig.suptitle('Vote Extensions Testnet - ' + release)

# Save the figure with subplots
fig.savefig(os.path.join(path,'all_experiments.png'))



#Group by configuration
groups = csv.groupby(['connections','rate'])

#number of rows and columns in the graph
ncols = 2 if groups.ngroups > 1 else 1
nrows = int( np.ceil(groups.ngroups / ncols)) if groups.ngroups > 1 else 1
fig, axes = plt.subplots(nrows=nrows, ncols=ncols, figsize=(6*ncols, 4*nrows), sharey=True)
fig.tight_layout(pad=5.0)

#Plot configurations as subplots 
for (key,ax) in zip(groups.groups.keys(), [axes] if ncols == 1 else axes.flatten()):
    group = groups.get_group(key)
    ax.set_ylabel('latency (s)')
    ax.set_xlabel('experiment time (s)')
    ax.grid(True)
    (con,rate) = key
    label = 'c='+str(con) + ' r='+ str(rate)
    ax.set_title(label)

    #Group by experiment 
    paramGroups = group.groupby(['experiment_id'])
    for (subKey) in paramGroups.groups.keys():
        subGroup = paramGroups.get_group(subKey)
        startTime = subGroup.block_time.min()
        subGroupMod = subGroup.block_time.apply(lambda x: x - startTime)
        ax.scatter(subGroupMod, subGroup.duration_ns, label=label)
    #ax.legend()
    

    #Save individual axes
    extent = ax.get_window_extent().transformed(fig.dpi_scale_trans.inverted())
    fig.savefig(os.path.join(path,'c'+str(con) + 'r'+ str(rate) + '.png'), bbox_inches=extent.expanded(1.2, 1.3))

fig.suptitle('Vote Extensions Testnet - ' + release)


# Save the figure with subplots
fig.savefig(os.path.join(path,'all_configs.png'))


fig, axes = plt.subplots(nrows=nrows, ncols=ncols, figsize=(6*ncols, 4*nrows), sharey=True)
fig.tight_layout(pad=5.0)

#Plot configurations as subplots 
for (key,ax) in zip(groups.groups.keys(), [axes] if ncols == 1 else axes.flatten()):
    group = groups.get_group(key)
    ax.set_ylabel('latency (s)')
    ax.set_xlabel('experiment time (s)')
    ax.grid(True)
    (con,rate) = key
    label = 'c='+str(con) + ' r='+ str(rate)
    ax.set_title(label)

    #Group by experiment, but merge them as a single experiment
    paramGroups = group.groupby(['experiment_id'])
    for (subKey) in paramGroups.groups.keys():
        subGroup = paramGroups.get_group(subKey)
        startTime = subGroup.block_time.min()
        subGroupMod = subGroup.block_time.apply(lambda x: x - startTime)
        ax.scatter(subGroupMod, subGroup.duration_ns, marker='o',c='#1f77b4')
    
    #Save individual axes
    extent = ax.get_window_extent().transformed(fig.dpi_scale_trans.inverted())
    (con,rate) = key
    fig.savefig(os.path.join(path,'c'+str(con) + 'r'+ str(rate) + '_merged.png'), bbox_inches=extent)

plt.show()
