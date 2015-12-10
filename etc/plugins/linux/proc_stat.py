from plugin import BasePlugin

import re

class ProcStat(BasePlugin):

	OldValues = {}

	def run(self, event):
		with open('/proc/stat', 'r') as f:
			for line in f:
				if re.search('^processes', line):
					data = line.split()
					service = [ "graphite:os.processes.forkrate", "opentsdb:os.processes.forkrate", "zabbix:os.processes.forkrate" ]
					event.send(service, float(data[1]), diff=True)
				if re.search('(^cpu)', line):
					m = re.match('^cpu(\d+|\s)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)', line)
					number = m.group(1)
					if number == ' ':
						number = 'total'
					u2, n2, s2, i2 = int(m.group(2)), int(m.group(3)), int(m.group(4)), int(m.group(5))
					if number in self.OldValues:
						u1, n1, s1, i1 = self.OldValues[number]
						used = (u2 + n2 + s2) - (u1 + n1 +s1)
						total = used + (i2 - i1)
						metric = used / total
						service = ["graphite:os.cpu.usage.{0}".format(number),"opentsdb:os.cpu.usage","zabbix:os.cpu.usage[{0}]".format(number)]
						event.send(service,metric,tags={"cpu":"{0}".format(number)})
					self.OldValues[number] = [u2, n2, s2, i2]


