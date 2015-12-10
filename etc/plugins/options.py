import optparse, socket

class Args():

	def __init__(self):
		parser = optparse.OptionParser()
		parser.add_option("-u", "--url", default="127.0.0.1:2003")
		parser.add_option("-s", "--server", default=[], action="append")
		parser.add_option("--add-tag-fqdn", default=True, help="Auto tag 'FQDN'")
		parser.add_option("--add-revert-graphite-fqdn", default=True, help="Add revert fqdm prefix for graphite")
		parser.add_option("-f", "--set-fqdn", default=socket.gethostname())
		self.args, _ = parser.parse_args()

		# set canonical fqdn
		self.args.__dict__["fqdn"] = self.args.__dict__["set_fqdn"]

	def __getattr__(self, name):
		try:
			return self.args.__dict__[name]
		except KeyError:
			return None
