from plugin import BasePlugin

import re

class DiskStats(BasePlugin):
	def run(self, event):
		with open('/proc/diskstats', 'r') as f:
			for line in f:
				if re.search('(ram|loop\d+)', line):
					continue
				m = re.match('^(?:\s+\d+){2}\s+([\w\d]+) (.*)$', line)
				if m is None:
					continue
				dev, val = m.group(1), m.group(2)
				if re.search('\d+', dev):
					continue
				val = [int(x) for x in val.split()]
				read, write = val[0], val[4]
				iops = read + write

				service = ["graphite:os.disk.{0}.{1}".format(dev,"read"), "opentsdb:os.disk.iops", "zabbix:os.disk.read[{0}]".format(dev)]
				event.send(service, read, tags = {"dev":dev, "operation":"read"},  diff=True)

				service = ["graphite:os.disk.{0}.{1}".format(dev,"write"), "opentsdb:os.disk.iops", "zabbix:os.disk.write[{0}]".format(dev)]
				event.send(service, write, tags ={"disk":dev, "operation":"write"}, diff=True)

				service = ["graphite:os.disk.{0}.{1}".format(dev,"iops"), "zabbix:os.disk.iops[{0}]".format(dev)]
				event.send(service, iops, diff=True)
