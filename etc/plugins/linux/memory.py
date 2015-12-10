from plugin import BasePlugin

class Memory(BasePlugin):
	def run(self, event):
		meminfo = {}
		with open('/proc/meminfo', 'r') as f:
			for line in f:
				data = line.split()
				key, val = data[0], data[1]
				if key == "MemFree:":
					meminfo["free"] = float(val) * 1024
				elif key == "Cached:":
					meminfo["cached"] = float(val) * 1024
				elif key == "Buffers:":
					meminfo["buffers"] = float(val) * 1024
				elif key == "MemTotal:":
					meminfo["total"] = float(val) * 1024
				elif key == "SwapTotal:":
					meminfo["swap_total"] = float(val) * 1024
				elif key == "SwapFree:":
					meminfo["swap_free"] = float(val) * 1024
		meminfo["used"] = meminfo["total"] - meminfo["free"]
		meminfo["free_bc"] = meminfo["free"] + meminfo["buffers"] + meminfo["cached"]
		meminfo["free_fraction"] = (meminfo["free_bc"] / meminfo["total"]) * 100
		meminfo["swap_fraction"] = 0
		if meminfo["swap_total"] != 0:
			meminfo["swap_fraction"] = (1 - (meminfo["swap_free"] / meminfo["swap_total"])) * 100
		for info in ["free_fraction", "swap_fraction", "free", "total", "cached", "buffers", "used", "free_bc"]:
			service = ["graphite:os.memory.{0}".format(info), "opentsdb:os.memory.{0}".format(info), "zabbix:os.memory[{0}]".format(info)]
			event.send(service, meminfo[info])
