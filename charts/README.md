
# Accessing clamav-rest installation

clamav-rest is available as application service, inside deployments.
That means, all the pods in deployment and namespace can reach the service.


Example:

Running curl from inside a pod in the same namespace as clamav-rest:
```
curl http://clamav-rest-staging:9000/

{"Pools":"1","State":"STATE: VALID PRIMARY","Threads":"THREADS: live 1  idle 0 max 10 idle-timeout 30","Memstats":"MEMSTATS: heap N/A mmap N/A used N/A free N/A releasable N/A pools 1 pools_used 1373.773M pools_total 1373.820M","Queue":"QUEUE: 0 items"}
```

There's also route defined as https (self signed) and FQDN address works as well from inside EF infrasrtructure:
```
curl -k https://clamav-rest-staging-open-vsx-org-staging.apps.okd-c1.eclipse.org/
```

Also, in order to access from committers workstations, configuration mentioned in https://gitlab.eclipse.org/eclipsefdn/it/releng/internal-services should be followed.


Please read README.md in the root of this repository for more information about accessing clamav-rest and appropriate URLs.



TODO:
- improve Dockerfile
    - install curl
- Move the repository to suitable location
- Move image to suitable container registry (ghcr.io/eclipsefdn ?)