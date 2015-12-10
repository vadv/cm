import time, threading, sys

class BasePlugin(object):

	Interval = 60

	def __init__(self, event, args = {}):
		self.event = event
		self.args = args
		self.thread = threading.Thread(target=self._loop)
		self.thread.daemon = True
		self._start_loop()

	def _loop(self):
		while(True):
			calc_time = time.time()
			self.run(self.event)
			calc_time =  self.Interval - (time.time() - calc_time)
			if calc_time > 0:
				time.sleep(calc_time)
			else:
				sys.exit("Error: executed time is too big")

	def _start_loop(self):
		self.thread.start()

	def isAlive(self):
		return self.thread.isAlive()
