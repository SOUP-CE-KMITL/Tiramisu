import os


name = raw_input("container name?\n>> ")
command = "docker stop " + name
os.system(command)

old_storage = raw_input("old storage?\n>> ")
new_storage = raw_input("new storage?\n>> ")
rbd_image_name = raw_input("rbd image name?\n>> ")
command = "rbd cp " + old_storage + "/" + rbd_image_name + " " + new_storage + "/" + rbd_image_name
os.system(command)

imagename = raw_input("image_name?\n>> ")
other_arg = raw_input("other argument?\n>> ")
command = "docker run -it --volume-driver=rbd --volume " + new_storage + "/" + rbd_image_name  + ":/mnt/foo/mysql " + other_arg + " -d " + imagename
print(command)
os.system(command)
