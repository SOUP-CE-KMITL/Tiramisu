import os
import subprocess
import psycopg2
import sys

try:
	conn = psycopg2.connect(database='tiramisu', user='postgres', host='localhost', port='5432', password='12344321')
except:
	print "Nooooooooo"

c = conn.cursor()

arg = sys.argv
name = arg[1]

c.execute("select status from tiramisu_vm where name=%s", (name,))
status = c.fetchone()
if status[0] == 1:
	command = "sudo virsh shutdown " + name
	os.system(command)
	print(command)
	print "########## shutting down ##########"
	while True:
		p = subprocess.Popen(['sudo', './automated_move/check_shutdown_vm', name], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
		status, err = p.communicate()
		if status == "shut":
			print "########## shut down complete ##########"
			c.execute("update tiramisu_vm set status=0 where name=%s", (name,))
			break
		else:
			command = "sudo virsh shutdown " + name
			os.system(command)
			print(command)

else:
	print "already shut down"

command = "sudo virsh undefine " + name
print(command)
os.system(command)

c.execute("select * from tiramisu_storage where vm_name=%s",(name,))
data = c.fetchone()
old_pool = data[1]
command = "sudo rm -f image/" + old_pool + "/" + name
print(command)
os.system(command)

command = "sudo rm -f image/config/" + name + ".xml"
print(command)
os.system(command)
