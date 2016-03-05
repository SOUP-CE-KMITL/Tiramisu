import os
import subprocess
import psycopg2
import sys

if __name__ == "__main__":
	try:
	    conn = psycopg2.connect(database='tiramisu', user='postgres', host='localhost', port='5432', password='12344321')
	except:
	    print "Nooooooooo"

	c = conn.cursor()

	arg 			= sys.argv
	name 			= arg[1]
	name_display	= arg[2]
	id_owner		= arg[3]

	command = "cp ../image/HDD/default ../image/HDD/" + name
	print(command)
	os.system(command)

	command = "cp ../image/config/default.xml ../image/config/" + name + ".xml"
	print(command)
	os.system(command)

	command = "sed -i -e 's/<name>default<\/name>/<name>" + name + "<\/name>/g' ../image/config/" + name + ".xml"
	print(command)
	os.system(command)

	command = "sed -i -e 's/\/image\/HDD\/default/\/image\/HDD\/" + name + "/g' ../image/config/" + name + ".xml"
	print(command)
	os.system(command)

	command = "sudo virsh define ../image/config/" + name + ".xml"
	print(command)
	os.system(command)

	command = "sudo virsh start " + name
	print(command)
	os.system(command)

	print "########## Preparing machine ##########"

	while True:
	        p = subprocess.Popen(['sudo', './kvm_findip', name], stdout=subprocess.PIPE, stderr=subprocess.PIPE)
	        ip, err = p.communicate()
	        if ip != "":
	                print "########## Complete ##########"
	                print "Your IP : "
	                print(ip)
	                break

	print "Default username : centos7\nDefault password : root"

	command = "./config_staic_ip " + ip
	print(command)
	os.system(command)

	c.execute("insert into tiramisu_vm (owner,name,ip,status,size,cost,name_display) values (%s,%s,%s,1,8589.93,429.49,%s)", (id_owner,name,ip,name_display,))
	c.execute("insert into tiramisu_storage (vm_name,current_pool,appropiate_pool,notice) values (%s,'HDD','HDD',1)", (name,))
	conn.commit()
	c.close()
