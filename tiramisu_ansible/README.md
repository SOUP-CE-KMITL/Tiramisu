# Automated deploy ceph cluster

#### Deploy monitor node

```
ansible-playbook hp6_node.yml
```

#### Copy admin keyring from monitor node to othor node

```
ssh hp6
scp /etc/ceph/ceph.client.admin.kyring ceph@10.0.1.*:/home/ceph
```

#### Deploy OSD node
```
ansible-playbook hp7_node.yml dell1_node.yml
```

#### Deploy mds node and ceph file system

```
ansible-playbook hp1_node.yml --ask-pass
```
