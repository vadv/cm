from plugin import BasePlugin

import re, os

class DiskSize(BasePlugin):

	NotMonitFsType = ["cgroup", "securityfs", "debugfs", "sysfs", "devpts", "squashfs", "proc", "devtmpfs", "fuse.sshfs", "fuse.gvfsd-fuse"]

	def run(self, event):
		with open('/proc/mounts', 'r') as f:
			for line in f:
				data = line.split()
				point, fs = data[1], data[2]
				if fs in self.NotMonitFsType:
					continue
				stat = os.statvfs(point)

				human_point = point
				if human_point == "/":
					human_point = "root"
				human_point = re.sub('^\/','', human_point)
				human_point = re.sub('\/','_', human_point)
				human_point = re.sub('\-','_', human_point)

				service = ["graphite:os.mount.{0}.bytes_total".format(human_point), "opentsdb:os.mount.bytes_total", "zabbix:os.mount[{0},{1}]".format(human_point,"bytes_total")]
				event.send(service,stat.f_blocks*stat.f_bsize,tags={"point":human_point})

				service = ["graphite:os.mount.{0}.bytes_free".format(human_point), "opentsdb:os.mount.bytes_free", "zabbix:os.mount[{0},{1}]".format(human_point,"bytes_free")]
				event.send(service,stat.f_bfree*stat.f_bsize,tags={"point":human_point})

				if stat.f_blocks != 0:
					metric = 100 * (1 - (stat.f_blocks - stat.f_bavail) / stat.f_blocks )
					service = ["graphite:os.mount.{0}.bytes_percent_free".format(human_point), "opentsdb:os.mount.bytes_percent_free", "zabbix:os.mount[{0},{1}]".format(human_point,"bytes_percent_free")]
					event.send(service,metric,tags={"point":human_point})

				if stat.f_files != 0:
					metric = 100 * (1 - (stat.f_ffree - stat.f_favail) / stat.f_ffree )
					service = ["graphite:os.mount.{0}.inodes_percent_free".format(human_point), "opentsdb:os.mount.inodes_percent_free", "zabbix:os.mount[{0},{1}]".format(human_point,"inodes_percent_free")]
					event.send(service,metric,tags={"point":human_point})


