from plugin import BasePlugin

class La(BasePlugin):
	def run(self, event):
		with open('/proc/loadavg', 'r') as f:
			for line in f:
				data = line.split()
				service = ["graphite:os.la.one_min", "opentsdb:os.la.one_min", "zabbix:os.la.one_min"]
				event.send(service, float(data[0]))
