import os
import subprocess
import sqlite3


conn = sqlite3.connect('../tiramisu.db')
c = conn.cursor()

name = raw_input("# VM name\n>> ")

c.execute("select status from vm where name=?", (name,))
status = c.fetchone()
if status[0] == 1:
	command = "sudo virsh shutdown " + name
	os.system(command)
	print(command)
	print "########## shutting down ##########"
	while True:
        	p = subprocess.Popen(['sudo', './check_shutdown_vm', name], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        	status, err = p.communicate()
        	if status == "shut":
                	print "########## shut down complete ##########"
			c.execute("update vm set status=0 where name=?", (name,))
                	break

else:
	print "already shut down"
conn.commit()
c.close()
os.system("sudo virsh list --all")

