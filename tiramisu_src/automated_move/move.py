import os
import subprocess
import sys
import psycopg2

try:
    conn = psycopg2.connect(database='tiramisu', user='postgres', host='localhost', port='5432', password='12344321')
except:
    print "Nooooooooo"

c = conn.cursor()

arg = sys.argv
name = arg[1]

cost_mb_SSD = 0.090
cost_mb_HDD = 0.050

c.execute("select status from tiramisu_vm where name=%s", (name,))
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
            c.execute("update tiramisu_vm set status=0 where name=%s", (name,))
            break
        else:
            command = "sudo virsh shutdown " + name
            os.system(command)
            print(command)

else:
    print "already shut down"

os.system("sudo virsh list --all")

command = "sudo virsh undefine " + name
print(command)
os.system(command)

c.execute("select * from tiramisu_storage where vm_name=%s",(name,))
data = c.fetchone()
old_pool = data[1]
new_pool = data[2]
command1 = "rbd cp " + old_pool + "/" + name + " " + new_pool + "/" + name
command2 = "rbd rm " + old_pool + "/" + name
command = command1 + " && " + command2
print(command)
os.system(command)

command = "sed -i -e 's/" + old_pool + "\/" + name + "/" + new_pool + "\/" + name + "/g' ../image/config/" + name + ".xml"
print(command)
os.system(command)

command = "sudo virsh define ../image/config/" + name + ".xml"
print(command)
os.system(command)

command = "sudo virsh start " + name
print(command)
os.system(command)

c.execute("select size from tiramisu_vm where name=%s", (name,))
size = c.fetchone()
if new_pool=='HDD':
    cost = float(size[0]) * cost_mb_HDD
    c.execute("select latency_hdd from tiramisu_state where vm_name=%s", (name,))
    latency = c.fetchone()
    c.execute("select iops_hdd from tiramisu_state where vm_name=%s", (name,))
    iops = c.fetchone()
else:
    cost = float(size[0]) * cost_mb_SSD
    c.execute("select latency_ssd from tiramisu_state where vm_name=%s", (name,))
    latency = c.fetchone()
    c.execute("select iops_ssd from tiramisu_state where vm_name=%s", (name,))
    iops = c.fetchone()

print "########## Start complete ##########"


c.execute("update tiramisu_vm set status=1,cost=%s where name=%s",(cost,name,))
c.execute("update tiramisu_storage set current_pool=%s where vm_name=%s",(new_pool,name,))
c.execute("update tiramisu_state set latency=%s,iops=%s where vm_name=%s",(latency,iops,name))
conn.commit()
c.close()
