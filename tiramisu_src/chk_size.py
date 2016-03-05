import subprocess, sys

p = subprocess.Popen(['df', '-h','grep','/dev/mapper/centos-home'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
status, err = p.communicate()
data = status.split()
percent_hdd = data[11][:-1]

p = subprocess.Popen(['df', '-h','grep','/dev/sdb1'], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
status, err = p.communicate()
data = status.split()
percent_ssd = data[11][:-1]

if int(percent_hdd) >= 80 or int(percent_ssd) >= 80:
	print("FULL")
	print "HDD:", percent_hdd, ", SSD:", percent_ssd
	sys.exit(1)
else:
	print("ok")
	print "HDD:", percent_hdd, ", SSD:", percent_ssd
	sys.exit(0)
