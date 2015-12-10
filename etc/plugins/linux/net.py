from plugin import BasePlugin

import re

class Net(BasePlugin):

	Filter = [ "rx bytes", "rx errs", "rx drop", "tx bytes", "tx errs", "tx drop" ]
	Words = [ "rx bytes", "rx packets", "rx errs", "rx drop", "rx fifo", "rx frame", "rx compressed", "rx multicast",
			"tx bytes", "tx packets", "tx errs", "tx drop", "tx fifo", "tx colls", "tx carrier", "tx compressed" ]

	def run(self, event):
		with open('/proc/net/dev', 'r') as f:
			for line in f:
				m = re.match('(\w*)\:\s*([\s\d]+)\s*', line)
				if m is None:
					continue
				iface = m.group(1)
				values = [int(x) for x in m.group(2).split()]
				for index, service in enumerate(self.Words):
					if not service in self.Filter:
						continue
					direction, typ = service.split()
					service = ["graphite:os.net.{0}.{1}".format(iface, typ), "opentsdb:os.net.{0}".format(typ), "zabbix:os.net[{0},{1},{2}]".format(iface, typ, direction)]
					metric = values[index]
					event.send(service, metric, tags={"interface":iface, "direction":direction}, diff=True)

