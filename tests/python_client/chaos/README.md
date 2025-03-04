# Chaos Tests
## Goal
Chaos tests are designed to check the reliability of Milvus.

For instance, if one pod is killed:
   - verify that it restarts automatically 
   - verify that the related operation fails, while the other operations keep working successfully during the absence of the pod
   - verify that all the operations work successfully after the pod back to running state
   - verify that no data lost

## Prerequisite
Chaos tests run in pytest framework, same as e2e tests. 

Please refer to [Run E2E Tests](https://github.com/milvus-io/milvus/blob/master/tests/README.md)

## Test Scenarios
### Milvus in cluster mode
1. root coordinator pod is killed
   
2. proxy pod is killed

3. data coordinator pod is killed

4. data node pod is killed

5. index coordinator pod is killed

6. index node pod is killed

7. query coordinator pod is killed

8. query node pod is killed

9. minio pod is killed

### Milvus in standalone mode
1. standalone pod is killed

2. minio pod is killed

## How it works
- Test scenarios are designed by different chaos objects
- Every chaos object is defined in one yaml file locates in  folder `chaos_objects`
- Every chaos yaml file specified by `ALL_CHAOS_YAMLS` in `constants.py` would be parsed as a parameter and be passed into `test_chaos.py`
- All expectations of every scenario are defined in `testcases.yaml` locates in folder `chaos_objects`
- [Chaos Mesh](https://chaos-mesh.org/) is used to inject chaos into Milvus in `test_chaos.py`

## Run
### Manually
Run a single test scenario manually(take query node pod is killed as instance):
1. update `ALL_CHAOS_YAMLS = 'chaos_querynode_podkill.yaml'` in `constants.py`

2. run the commands below:
```bash
cd /milvus/tests/python_client/chaos

pytest test_chaos.py --host ${Milvus_IP} -v
```

### Nightly 
still in planning 

### Todo
* pod_failure
* container_kill
* network attack