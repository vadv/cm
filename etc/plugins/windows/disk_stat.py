from plugin import BasePlugin

from windows.helpers import DiskInfo, PerfData

class DiskStat(BasePlugin):

	def run(self, event):

		for disk in DiskInfo.get_fixed_drivers():

			disk_usage = DiskInfo.get_drive_info(disk)

			service = ["graphite:os.mount.{0}.bytes_total".format(disk), "opentsdb:os.mount.bytes_total", "zabbix:os.mount[{0},{1}]".format(disk,"bytes_total")]
			event.send(service,disk_usage.total,tags={"point":disk})

			service = ["graphite:os.mount.{0}.bytes_free".format(disk), "opentsdb:os.mount.bytes_free", "zabbix:os.mount[{0},{1}]".format(disk,"bytes_free")]
			event.send(service,disk_usage.free,tags={"point":disk})


			perf_service = [r'\LogicalDisk(*:\\)\% Disk Read Time'.format(disk), r'\LogicalDisk(*:\\)\% Disk Write Time'.format(disk)]
			data = PerfData.get(perf_service, fmts='double', delay=10000)

			service = ["graphite:os.disk.{0}.{1}".format(disk,"time.read"), "opentsdb:os.disk.time", "zabbix:os.disk.[{0},{1}]".format(dev, "read_time")]
			event.send(service, data[0], tags ={"disk":disk, "operation":"read"})
			service = ["graphite:os.disk.{0}.{1}".format(disk,"time.write"), "opentsdb:os.disk.time", "zabbix:os.disk.[{0},{1}]".format(dev, "write_time")]
			event.send(service, data[1], tags ={"disk":disk, "operation":"write"})
