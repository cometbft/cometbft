#!/usr/bin/env python3

import subprocess
import asyncio
import os
import signal


def set_project_root():
	current_file_path = os.path.abspath(__file__)
	candidate = os.path.dirname(os.path.dirname(os.path.dirname(current_file_path)))
	if os.path.basename(candidate) == "cometbft":
		os.chdir(candidate)
		return
	raise FileNotFoundError("Can't locate project root")


async def read_process_output(process, node_started_event):
	while True:
		line = await process.stdout.readline()
		if not line:
			break
		ln = line.decode().strip()
		if "Started node" in ln:
			node_started_event.set()


async def clear_n_launch_localnet(node_started_event):
	clear_command = ["sudo", "rm", "-rf", "build/"]
	build_command = ["sudo", "make", "build-linux"]
	launch_command = ["sudo", "make", "localnet-start"]
	subprocess.run(clear_command, capture_output=True, text=True, check=True)
	subprocess.run(build_command, capture_output=True, text=True, check=True)
	process = await asyncio.create_subprocess_exec(*launch_command, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)
	asyncio.create_task(read_process_output(process, node_started_event))
		


def launch_payload(r, s, T):
	payload_script_path = os.path.join("test", "loadtime", "build", "load")
	payload_command = [payload_script_path, "-T", str(T), "-r", str(r), "-s", str(s),  "--endpoints", "ws://localhost:26657/websocket"]
	result = subprocess.run(payload_command, text=True, check=True, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE)
	storage_path = os.path.join("build", "node0", "data", "blockstore.db")
	measure_command = ["du", "-h", storage_path]
	result = subprocess.run(measure_command, capture_output=True, text=True, check=True)
	return result.stdout.split()[0]


def stop_localnet():
	stop_command = ["sudo", "make", "build", "localnet-stop"]
	subprocess.run(stop_command, capture_output=True, text=True, check=True)


async def run():
	tx_sizes = [2048, 3000, 4096, 5000, 8192, 10000]
	rates = [1300, 1500, 1700, 1900, 2100]
	T = 60
	for tx_size in tx_sizes:
		for rate in rates:
			node_started_event = asyncio.Event()
			localnet_process = asyncio.create_task(clear_n_launch_localnet(node_started_event))
			await node_started_event.wait()
			resulting_storage_size = launch_payload(rate, tx_size, T)
			print(f"r {rate}; s {tx_size}; T {T}; storage size: {resulting_storage_size}")
			stop_localnet()


def main():
	set_project_root()
	asyncio.run(run())


main()