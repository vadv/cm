import sys

try:
	import json
except ImportError:
	import simplejson as json

try:
	import http.client as httpclient
except ImportError:
	import httplib as httpclient


class Event():

	def __init__(self, newargs={}):
		self.args = newargs
		self.previous = {}

	def send(self, service, value, tags={}, fqdn=None, diff=False, url=None):

		if url is None:
			url = self.args.url
		if fqdn is None:
			fqdn = self.args.fqdn

		if diff == True:
			key = "{0}_{1}_{2}".format(str(service), fqdn, url)
			new_value = value
			try:
				value = value - self.previous[key]
			except KeyError:
				self.previous[key] = new_value
				return
			self.previous[key] = new_value

		if type(service) is list:
			for real_service in service:
				self._send(real_service, value, tags, fqdn, url)
			return

		self._send(service, value, tags, fqdn, url)

	def _get_service(self, service):
		server, service = service.split(":")
		return server, service

	def _send(self, service, value, tags, fqdn, url):
		server, service_name = self._get_service(service)
		if self.args.server:
			if not server in self.args.server:
				return
		conn = httpclient.HTTPConnection(url, timeout=1)

		# auto tag fqdn
		if self.args.add_tag_fqdn:
			if not "fqdn" in tags:
				tags["fqdn"] = fqdn

		# set graphite service
		if self.args.add_revert_graphite_fqdn and server == "graphite":
			graphite_fqdn = fqdn.split(".")
			graphite_fqdn = graphite_fqdn[::-1]
			service_name = "{0}.{1}".format(".".join(graphite_fqdn), service_name)

		data = json.dumps({
			"fqdn":fqdn,
			server:service_name,
			"value":value,
			"tags": tags,
		})
		conn.request("POST", "/", data)
		response = conn.getresponse()
		if response.status != 200:
			sys.exit("BAD HTTP STATUS: {0}".format(response.status))



