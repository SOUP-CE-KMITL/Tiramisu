# SystemTap script

### Install SystemTap

```
sudo yum install systemtap systemtap-runtime
```

Install kernel-debuginfo-common
```
wget http://debuginfo.centos.org/7/x86_64/kernel-debuginfo-common-x86_64-`uname -r`.rpm
```
```
sudo rpm -ihv kernel-debuginfo-common-x86_64-`uname -r`.rpm
```

Install kernel-debuginfo-
```
wget http://debuginfo.centos.org/7/x86_64/kernel-debuginfo-`uname -r`.rpm
```
```
sudo rpm -ihv kernel-debuginfo-`uname -r`.rpm
```

```
stap-prep
```

Optional
```
sudo yum install kernel-devel-(kernel_version)
```




### Run script

```
sudo stap name.stp
```


### Create module

Create .ko module for run at another (same kernel_version) machine

```
stap -r kernel_version script -m module_name
```
after run this command it will generate **mudule_name.ko**

### Run module at another machine

```
staprun module_name.ko
```
