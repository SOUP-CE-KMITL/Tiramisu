
#########################################################################################
# data recieve from outside

def requirements():
	# create requirement table to store all of this (column by column)
	# this function will read data from table and store in dictionary
	# edit value when user change requirement

	latency = 4
	latency_max = 6
	percentl = 5
	iops_min = 400
	iops = 500
	percenti = 10
	cost = 50
	cost_max = 90
	percentc = 10
	app_type = 1

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

def cube(choice):
	# create cube table to store all of this (column by column) (copy from requirement at first)
	# this function will read data from table and store in dictionary
	# edit value when squeeze or puff
	if choice == 1 or choice == 2 or choice == 3:
		latency_min = 3.8 
		latency = 4
		latency_max = 6
		percentl = 5
		iops_min = 400
		iops = 500
		iops_max = 550
		percenti = 10
		cost_min = 45
		cost = 50
		cost_max = 90
		percentc = 10
		app_type = 1

	else:
		latency_min = 3.6
		latency = 4
		latency_max = 6.2
		percentl = 5
		iops_min = 350
		iops = 500
		iops_max = 600
		percenti = 10
		cost_min = 40
		cost = 50
		cost_max = 95
		percentc = 10
		app_type = 1		

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

def storage_status(choice):
	# recieve from systemtap
	# systemtap will exec this code

	if choice == 1:
		latency_hdd = 5
		iops_hdd = 401
		latency_ssd = 2
		iops_ssd = 600
	elif choice == 2:
		latency_hdd = 8
		iops_hdd = 401
		latency_ssd = 3.7
		iops_ssd = 570
	elif choice == 3:
		latency_hdd = 6
		iops_hdd = 500
		latency_ssd = 2
		iops_ssd = 1000
	elif choice == 4:
		latency_hdd = 6.1
		iops_hdd = 500
		latency_ssd = 4
		iops_ssd = 600
	else:
		latency_hdd = 6.1
		iops_hdd = 500
		latency_ssd = 3.9
		iops_ssd = 500

	return {	"latency_hdd" 	: latency_hdd,
				"latency_ssd" 	: latency_ssd,
				"iops_hdd"		: iops_hdd,
				"iops_ssd"		: iops_ssd }
	
def get_size_vm():
	# receive from outside
	size = 1000
	return size

def read_current_storage():
	# read from DB
	return 'HDD'

#########################################################################################

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
	cost_mb_SSD = 0.090
	cost_mb_HDD = 0.050

	print("""select case:\n1. current in requirements
2. current out of cube and another storage closer than current
3. current out of cube and current closest
4. current out of small cube(or requirements) but in cube and another storage is same
5. current out of small cube(or requirements) but in cube and another storage in small cube(or requirements)""")
	choice = input(">>> ")

	requirements = requirements()
	cube = cube(choice)

	get_size_vm = get_size_vm()
	cost_SSD = cal_cost(cost_mb_SSD, get_size_vm)
	cost_HDD = cal_cost(cost_mb_HDD, get_size_vm)

	storage_status = storage_status(choice)

	point_storage = { 	"SSD" : [storage_status["latency_ssd"], storage_status["iops_ssd"], cost_SSD],
						"HDD" : [storage_status["latency_hdd"], storage_status["iops_hdd"], cost_HDD] }

	current = read_current_storage()
	current_point = point_storage[current]

	if is_in_cube(cube, current_point):
		print squeeze(point_storage, cube, current_point, current)
	else:
		# current storage not meet requirement
		result = find_other_storage_in_cube(point_storage, cube, current_point, current)
		if result[0] == 1:
			# still have other point in requirement
			current_point = result[1]
			current = result[2]
			print squeeze(point_storage, cube, current_point, current)
		else:
			# not have any point in requirement
			print puff(point_storage, cube, current)
