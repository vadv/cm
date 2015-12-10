from plugin import BasePlugin

from windows.helpers import PerfData

class Cpu(BasePlugin):

	def run(self, event):
		# todo: получить список процессоров.
		# вообще вся эта информация есть в клиенте zabbix
		# но! есть возможность получить PerfData для процесса по маске
		# поэтому в дальнейшем этот плагин будет дополнен конкретикой по процессам postgresql
		cpu_number = '*'
		data = PerfData.get(r'\Processor({0})\% Processor Time'.format(cpu_number), fmts='double', delay=10000)
		if cpu_number = '*':
			cpu_number = 'total'
		service = ["graphite:os.cpu.usage.{0}".format(cpu_number),"opentsdb:os.cpu.usage","zabbix:os.cpu.usage[{0}]".format(cpu_number)]
		event.send(service,data[0],tags={"cpu":"{0}".format(cpu_number)})
