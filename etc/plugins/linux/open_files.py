from plugin import BasePlugin

class OpenFiles(BasePlugin):

	def run(self, event):
		with open('/proc/sys/fs/file-nr', 'r') as f:
			for line in f:
				metric = int(line.split()[0])
				service = ["graphite:os.openfiles", "opentsdb:os.openfiles", "zabbix:os.openfiles"]
				event.send(service, metric)
