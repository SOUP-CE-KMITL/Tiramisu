import os

imagename = raw_input("Image name? ")
command = "docker run -it --volume-driver=rbd --volume foo:/mnt/foo -d " + imagename + " bash"
os.system(command)
