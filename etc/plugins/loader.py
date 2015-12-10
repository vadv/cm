#!/usr/bin/env python

import time, sys
import utils, options
from plugin import BasePlugin

if sys.platform == "linux" or sys.platform == "linux2":
	from linux import *

if sys.platform == "win32" or sys.platform == "win64":
	from windows import *

Args = options.Args()
Event = utils.Event(Args)
Plugins = []

for plugin in BasePlugin.__subclasses__():
	Plugins.append(plugin(Event, Args))

while True:
	time.sleep(1)
	for plugin in Plugins:
		if not plugin.isAlive():
			sys.exit("Error: One or more of monitoring thread is died.")
