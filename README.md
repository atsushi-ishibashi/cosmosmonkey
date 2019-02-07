# cosmosmonkey
CLI to terminate EC2 instance which belongs to ECS Cluster.
Drains cluster instance, waits for a while and terminates the instance.

## Install
```
go get -u github.com/atsushi-ishibashi/cosmosmonkey/cmd/cosmosmonkey
```

## Usage
```
$ cosmosmonkey -h
Usage of ./cosmosmonkey:
  -cluster string
    	cluster name
  -instance string
    	instance id
  -max-drain-wait int
    	max wait time until draining finish (default 100)
```
