from plugin import BasePlugin

from windows.helpers import PerformanceInfo

class PerfInfo(BasePlugin):

	def run(self, event):

		perinfo = PerformanceInfo.get()

		service = ["graphite:os.memory.total", "opentsdb:os.memory.total", "zabbix:os.memory[total]"]
		event.send(service, perinfo.PhysicalTotal * perinfo.PageSize)

		service = ["graphite:os.memory.free", "opentsdb:os.memory.free", "zabbix:os.memory[free]"]
		event.send(service, perinfo.PhysicalAvailable * perinfo.PageSize)

		service = ["graphite:os.memory.cached", "opentsdb:os.memory.cached", "zabbix:os.memory[cached]"]
		event.send(service, perinfo.SystemCache * perinfo.PageSize)

		service = ["graphite:os.openfiles", "opentsdb:os.openfiles", "zabbix:os.openfiles"]
		event.send(service, perinfo.HandleCount)


