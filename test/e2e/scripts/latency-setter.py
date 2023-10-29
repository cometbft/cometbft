#!/usr/bin/python3
#
# Modified version of latsetter.py script from 
# https://github.com/paulo-coelho/latency-setter

import csv
import os
import sys
import netifaces as nif

def usage():
    print('Usage: \t', sys.argv[0], 'set ip-list latency-list iface-name')
    print('\tOR:\t', sys.argv[0], 'unset iface-name')
    exit()

if len(sys.argv) < 3:
    usage()

if sys.argv[1] == 'unset':
    command = 'tc qdisc del dev ' + sys.argv[2] + ' root'
    os.system(command)
    exit()

if sys.argv[1] != 'set':
    usage()

#reads IP/Zone file
ips = open(sys.argv[2], 'r')
ipD = csv.DictReader(ips)
ipData = []

for r in ipD:
    ipData.append(r)

#gets my address
iface = sys.argv[4]
myip = nif.ifaddresses(iface)[nif.AF_INET][0]['addr']

#reads latency file
with open(sys.argv[3], 'r') as f:
    reader = csv.reader(f)
    lats = list(list(rec) for rec in csv.reader(f, delimiter=',')) #reads csv into a list of lists

myzone = ''
for item in ipData:
    if item['IP'] == myip:
        myzone = item['Zone']

if myzone == '':
    print('IP', myip, 'not in IP file')
    exit()

azs = lats[0]
tlats = {}
for l in lats[1:]:
    for i in range(1, len(azs)):
        key = (l[0], azs[i])
        value = float(l[i])
        if value != 0:
            tlats[key] = value

print('# Setting rules for interface', iface, 'in zone', myzone, 'with IP', myip)

command = 'tc qdisc del dev ' + iface + ' root'
os.system(command)
command = 'tc qdisc add dev ' + iface + ' root handle 1: htb default 10'
os.system(command)
command = 'tc class add dev ' + iface + ' parent 1: classid 1:1 htb rate 1gbit 2> /dev/null'
os.system(command)
command = 'tc class add dev ' + iface + ' parent 1:1 classid 1:10 htb rate 1gbit 2> /dev/null'
os.system(command)
command = 'tc qdisc add dev ' + iface + ' parent 1:10 handle 10: sfq perturb 10'
os.system(command)

nextHandle = 11
for az in azs[1:]:
    lat = tlats.get((myzone, az))
    if lat == None:
        continue
    if lat > 0:#az != myzone:
        lat = tlats.get((myzone, az))
        delta = .05 * lat
        command = 'tc class add dev ' + iface + ' parent 1:1 classid 1:' + str(nextHandle) + ' htb rate 1gbit 2> /dev/null'
        os.system(command)
        command = 'tc qdisc add dev ' + iface + ' parent 1:' + str(nextHandle) + ' handle ' + str(
            nextHandle) + ': netem delay ' + str(lat) + 'ms ' + str(delta) + 'ms distribution normal'
        os.system(command)
        for item in ipData:
            if item['Zone'] == az:
                ip = item['IP']
                print(f'# Configuring latency from {myip} ({myzone}) to {ip} ({az}) set to {lat}ms +/- {delta}ms')
                command = 'tc filter add dev ' + iface + ' protocol ip parent 1: prio 1 u32 match ip dst ' + ip + '/32 flowid 1:' + str(
                    nextHandle)
                os.system(command)
        nextHandle += 1
