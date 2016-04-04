import psycopg2
import sys

#########################################################################################
# data recieve from outside

def requirements(name, c):
	# create requirement table to store all of this (column by column)
	# this function will read data from table and store in dictionary
	# edit value when user change requirement
	c.execute("select * from tiramisu_requirements where vm_name=%s", (name,))
	requirements = c.fetchone()

	latency 	= requirements[1]
	latency_max = requirements[2]
	percentl 	= requirements[3]
	iops_min 	= requirements[4]
	iops 		= requirements[5]
	percenti 	= requirements[6]
	cost 		= requirements[7]
	cost_max 	= requirements[8]
	percentc 	= requirements[9]
	app_type 	= requirements[10]

	return { 	"latency" 		: latency,
				"latency_max" 	: latency_max,
				"percentl" 		: percentl,
				"iops_min" 		: iops_min,
				"iops" 			: iops,
				"percenti" 		: percenti,
				"cost"			: cost,
				"cost_max" 		: cost_max,
				"percentc"		: percentc,
				"app_type"		: app_type }

def cube(name, c):
	# create cube table to store all of this (column by column) (copy from requirement at first)
	# this function will read data from table and store in dictionary
	# edit value when squeeze or puff

	c.execute("select * from tiramisu_cube where vm_name=%s", (name,))
	cube = c.fetchone()

	latency_min = cube[1]
	latency 	= cube[2]
	latency_max = cube[3]
	percentl 	= cube[4]
	iops_min 	= cube[5]
	iops 		= cube[6]
	iops_max 	= cube[7]
	percenti 	= cube[8]
	cost_min 	= cube[9]
	cost 		= cube[10]
	cost_max 	= cube[11]
	percentc 	= cube[12]
	app_type 	= cube[13]	

	return { 	"latency_min"	: latency_min,
				"latency" 		: latency,
				"latency_max" 	: latency_max,
				"percentl" 		: percentl,
				"iops_min" 		: iops_min,
				"iops" 			: iops,
				"iops_max"		: iops_max,
				"percenti" 		: percenti,
				"cost_min"		: cost_min,
				"cost"			: cost,
				"cost_max" 		: cost_max,
				"percentc"		: percentc,
				"app_type"		: app_type }

def get_state(name, c):
	c.execute("select * from tiramisu_state where vm_name=%s", (name,))
	state = c.fetchone()
	latency_vm 	= state[1]
	iops_vm 	= state[2]
	latency_hdd	= state[3]
	iops_hdd 	= state[4]
	latency_ssd	= state[5]
	iops_ssd 	= state[6]

	return {	"latency_vm"	: latency_vm,
				"iops_vm"		: iops_vm,
				"latency_hdd"	: latency_hdd,
				"iops_hdd"		: iops_hdd,
				"latency_ssd"	: latency_ssd,
				"iops_ssd"		: iops_ssd }

def cal_percent(pc, data):
	return (data * pc) / 100.00

def cal_cost(cost_mb, size):
	return cost_mb * size

def increase_size(cube, requirements):
	cube["percentl"] = cube["percentl"] + requirements["percentl"]
	cube["percenti"] = cube["percenti"] + requirements["percenti"]
	cube["percentc"] = cube["percentc"] + requirements["percentc"]

	cube["latency_min"] -= cal_percent(requirements["percentl"], cube["latency"])
	cube["latency_max"] += cal_percent(requirements["percentl"], cube["latency"])
	cube["iops_min"] -= cal_percent(requirements["percenti"], cube["iops"])
	cube["iops_max"] += cal_percent(requirements["percenti"], cube["iops"])
	cube["cost_min"] -= cal_percent(requirements["percentc"], cube["cost"])
	cube["cost_max"] += cal_percent(requirements["percentc"], cube["cost"])
	return cube

def decrease_size(cube, requirements):
	if cube["percentl"] > 0:
		cube["percentl"] = cube["percentl"] - requirements["percentl"]
	if cube["percenti"] > 0:
		cube["percenti"] = cube["percenti"] - requirements["percenti"]
	if cube["percentc"] > 0:
		cube["percentc"] = cube["percentc"] - requirements["percentc"]
	
	cube["latency_min"] += cal_percent(requirements["percentl"], cube["latency"])
	cube["latency_max"] -= cal_percent(requirements["percentl"], cube["latency"])
	cube["iops_min"] += cal_percent(requirements["percenti"], cube["iops"])
	cube["iops_max"] -= cal_percent(requirements["percenti"], cube["iops"])
	cube["cost_min"] += cal_percent(requirements["percentc"], cube["cost"])
	cube["cost_max"] -= cal_percent(requirements["percentc"], cube["cost"])
	return cube

def is_in_cube(cube, current_point):
	if 	(cube["latency_min"] <= current_point[0] and current_point[0] <= cube["latency_max"] and 
		cube["iops_min"] <= current_point[1] and current_point[1] <= cube["iops_max"] and 
		cube["cost_min"] <= current_point[2] and current_point[2] <= cube["cost_max"]) :
		return 1
	else:
		return 0

def find_other_storage_in_cube(point_storage, cube, current_point, current):
	first = 1
	found = 0
	for storage in point_storage:
		if storage != current:
			other_point = point_storage[storage]
			if is_in_cube(cube, other_point):
				if not found:
					found = 1
				if first:
					current_point = point_storage[storage]
					current = storage
					first = 0
				else:
					if cube["app_type"] == 1: 
						# website care low latency
						if other_point[0] < current_point[0]:
							current_point = other_point
							current = storage
					elif cube["app_type"] == 2:
						# database care high iops
						if other_point[1] > current_point[1]:
							current_point = other_point
							current = storage
					else:
						# batch job care low cost
						if other_point[2] < current_point[2]:
							current_point = other_point
							current = storage
	if found:
		return [1, current_point, current]
	else:
		return [0, [], 0]

def squeeze(point_storage, cube, current_point, current):
	while 1:
		cube = decrease_size(cube, requirements)
		if not is_in_cube(cube, current_point) :
			result = find_other_storage_in_cube(point_storage, cube, current_point, current)
			if result[0] == 1:
				current_point = result[1]
				current = result[2]
			else:
				return current
		elif cube["percentc"] <= 0 and cube["percenti"] <= 0 and cube["percentl"] <= 0:
			return current

def puff(point_storage, cube, current):
	while 1:
		cube = increase_size(cube, requirements)
		first = 1
		found = 0
		for storage in point_storage:
			other_point = point_storage[storage]
			if is_in_cube(cube, other_point):
				if not found:
					found = 1
				if first:
					current_point = point_storage[storage]
					current = storage
					first = 0
				else:
					if cube["app_type"] == 1: 
						# website care low latency
						if other_point[0] < current_point[0]:
							current_point = other_point
							current = storage
					elif cube["app_type"] == 2:
						# database care high iops
						if other_point[1] > current_point[1]:
							current_point = other_point
							current = storage
					else:
						# batch job care low cost
						if other_point[2] < current_point[2]:
							current_point = other_point
							current = storage
		if found:
			return current		

if __name__ == "__main__":
	try:
		conn = psycopg2.connect(database='tiramisu', user='postgres', host='localhost', port='5432', password='12344321')
	except:
		print "Nooooooooo"

	c = conn.cursor()

	cost_mb_SSD = 0.090
	cost_mb_HDD = 0.050

	arg 		= sys.argv
	name 		= arg[1]

	get_state = get_state(name, c)
	latency_vm 	= get_state["latency_vm"]
	iops_vm 	= get_state["iops_vm"]
	latency_hdd	= get_state["latency_hdd"]
	iops_hdd 	= get_state["iops_hdd"]
	latency_ssd	= get_state["latency_ssd"]
	iops_ssd 	= get_state["iops_ssd"]

	requirements = requirements(name, c)
	cube = cube(name, c)

	if (cube["latency"] == 0.0):
		cube["latency"] = 1

	if (cube["iops"] == 0.0):
                cube["iops"] = 1

	if (cube["cost"] == 0.0):
                cube["cost"] = 1

	

	c.execute("select * from tiramisu_vm where name=%s", (name,))
	vm_details = c.fetchone()
	c.execute("select * from tiramisu_storage where vm_name=%s", (name,))
	storage_vm = c.fetchone()

	get_size_vm = vm_details[5]
	cost_SSD = cal_cost(cost_mb_SSD, get_size_vm)
	cost_HDD = cal_cost(cost_mb_HDD, get_size_vm)

	if storage_vm[1] == 'SSD':
		point_storage = { 	"SSD" : [latency_vm, iops_vm, cost_SSD],
							"HDD" : [latency_hdd, iops_hdd, cost_HDD] }
	else:
		point_storage = { 	"SSD" : [latency_ssd, iops_ssd, cost_SSD],
							"HDD" : [latency_vm, iops_vm, cost_HDD] }

	c.execute("select current_pool from tiramisu_storage where vm_name=%s", (name,))
	pool = c.fetchone()
	current = pool[0]
	current_point = point_storage[current]
	
	if is_in_cube(cube, current_point):
		ans = squeeze(point_storage, cube, current_point, current)
	else:
		# current storage not meet requirement
		result = find_other_storage_in_cube(point_storage, cube, current_point, current)
		if result[0] == 1:
			# still have other point in requirement
			current_point = result[1]
			current = result[2]
			ans = squeeze(point_storage, cube, current_point, current)
		else:
			# not have any point in requirement
			ans = puff(point_storage, cube, current)

	if ans != storage_vm[2]:
		c.execute("update tiramisu_cube set latency_min=%s,latency=%s,latency_max=%s,percentl=%s,iops_min=%s,iops=%s,iops_max=%s,percenti=%s,cost_min=%s,cost=%s,cost_max=%s,percentc=%s,app_type=%s where vm_name=%s",(cube["latency_min"],cube["latency"],cube["latency_max"],cube["percentl"],cube["iops_min"],cube["iops"],cube["iops_max"],cube["percenti"],cube["cost_min"],cube["cost"],cube["cost_max"],cube["percentc"],cube["app_type"],name,))
		c.execute("update tiramisu_storage set appropiate_pool=%s where vm_name=%s",(ans,name,))
	conn.commit()
	c.close()
