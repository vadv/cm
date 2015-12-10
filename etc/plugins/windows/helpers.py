import ctypes, string, sys
from ctypes import (Structure, Union, WinError, byref, c_double, c_longlong,
					c_ulong, c_ulonglong, c_size_t, sizeof)
from ctypes.wintypes import HANDLE, LONG, LPCSTR, LPCWSTR, DWORD
from collections import namedtuple

HQUERY = HCOUNTER = HANDLE
psapi = ctypes.windll.psapi
pdh = ctypes.windll.pdh
kernel32 = ctypes.windll.kernel32

PDH_FMT_RAW     = 16L
PDH_FMT_ANSI    = 32L
PDH_FMT_UNICODE = 64L
PDH_FMT_LONG    = 256L
PDH_FMT_DOUBLE  = 512L
PDH_FMT_LARGE   = 1024L
PDH_FMT_1000    = 8192L
PDH_FMT_NODATA  = 16384L
PDH_FMT_NOSCALE = 4096L

class PerformanceInfo(Structure):
	_fields_ = [
		('size', c_ulong),
		('CommitTotal', c_size_t),
		('CommitLimit', c_size_t),
		('CommitPeak', c_size_t),
		('PhysicalTotal', c_size_t),
		('PhysicalAvailable', c_size_t),
		('SystemCache', c_size_t),
		('KernelTotal', c_size_t),
		('KernelPaged', c_size_t),
		('KernelNonpaged', c_size_t),
		('PageSize', c_size_t),
		('HandleCount', c_ulong),
		('ProcessCount', c_ulong),
		('ThreadCount', c_ulong),
	]
	def __init__(self):
		self.size = sizeof(self)
		super(PerformanceInfo, self).__init__()

	@classmethod
	def get(cls):
		perfinfo = PerformanceInfo()
		psapi.GetPerformanceInfo(byref(perfinfo), perfinfo.size)
		return perfinfo

class DiskInfo(object):

	Types = {
		0: 'UNKNOWN',
		1: 'NO_ROOT_DIR',
		2: 'REMOVABLE',
		3: 'FIXED',
		4: 'REMOTE',
		5: 'CDROM',
		6: 'RAMDISK',
	}

	@classmethod
	def get_fixed_drivers(cls):
		drives = []
		bitmask = kernel32.GetLogicalDrives()
		for letter in string.uppercase:
			if bitmask & 1:
				if kernel32.GetDriveTypeA(letter+':\\') == 3:
					drives.append(letter)
			bitmask >>= 1

		return drives

	@classmethod
	def get_drive_info(cls, drive):
		_diskusage = namedtuple('disk_usage', 'total used free')
		used, total, free = c_ulonglong(), c_ulonglong(), c_ulonglong()
		drive = drive + ':\\'
		ret = kernel32.GetDiskFreeSpaceExA(str(drive), byref(used), byref(total), byref(free))
		if ret == 0:
			raise
		else:
			return _diskusage(total.value, used, free.value)

class PDH_Counter_Union(Union):
	_fields_ = [
		('longValue', LONG),
		('doubleValue', c_double),
		('largeValue', c_longlong),
		('ansiValue', LPCSTR),
		('unicodeValue', LPCWSTR)
	]

class PDHFmtCounterValue(Structure):
	_fields_ = [
		('CStatus', DWORD),
		('union', PDH_Counter_Union),
	]

class PerfData(object):

	@classmethod
	def get(cls, counters, fmts='long', english=True, delay=0):

		if type(counters) is list:
			counters = [ unicode(c)  for c in counters ]
		else:
			counters = [ unicode(counters) ]

		getfmt = lambda fmt: globals().get('PDH_FMT_' + fmt.upper(), PDH_FMT_LONG)
		if type(fmts) is list:
			ifmts = [ getfmt(fmt)  for fmt in fmts ]
		else:
			ifmts = [getfmt(fmts)]
			fmts  = [fmts]
			if english:
				addfunc = pdh.PdhAddEnglishCounterW
			else:
				addfunc = pdh.PdhAddCounterW

		hQuery = HQUERY()
		hCounters = []
		values = []

		errs = pdh.PdhOpenQueryW(None, 0, byref(hQuery))
		if errs:
			raise WindowsError, 'PdhOpenQueryW failed, error: {0}'.format(errs)

		for counter in counters:
			hCounter = HCOUNTER()
			errs = addfunc(hQuery, counter, 0, byref(hCounter))
			if errs:
				raise WindowsError, 'PdhAddCounterW failed, error: {0}'.format(errs)
			hCounters.append(hCounter)

		errs = pdh.PdhCollectQueryData(hQuery)
		if errs:
			raise WindowsError, 'PdhCollectQueryData failed, error: {0}'.format(errs)
		if delay:
			kernel32.Sleep(delay)
			errs = pdh.PdhCollectQueryData(hQuery)
			if errs:
				raise WindowsError, 'PdhCollectQueryData failed, error: {0}'.format(errs)

		for i, hCounter in enumerate(hCounters):
			value = PDHFmtCounterValue()
			errs = pdh.PdhGetFormattedCounterValue(hCounter, ifmts[i], None, byref(value))
			if errs:
				raise WindowsError, 'PdhGetFormattedCounterValue failed, error: {0}'.format(errs)
			values.append(value)

		errs = pdh.PdhCloseQuery(hQuery)
		if errs:
			raise WindowsError, 'PdhCloseQuery failed, error: {0}'.format(errs)

		return tuple([ getattr(value.union, fmts[i] + 'Value')
				for i, value in enumerate(values) ])
