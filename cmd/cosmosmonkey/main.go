package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/atsushi-ishibashi/cosmosmonkey/model"
	"github.com/atsushi-ishibashi/cosmosmonkey/svc"
)

var (
	cluster      = flag.String("cluster", "", "cluster name")
	instance     = flag.String("instance", "", "instance id")
	maxDrainWait = flag.Int64("max-drain-wait", 100, "max wait time until draining finish")
)

func main() {
	flag.Parse()
	if *cluster == "" {
		log.Fatal("cluster required")
	}
	if *instance == "" {
		log.Fatal("instance required")
	}
	region := os.Getenv("AWS_REGION")
	defaultRegion := os.Getenv("AWS_DEFAULT_REGION")
	if region != "" {
		os.Setenv("_CM_AWS_REGION", region)
	} else if defaultRegion != "" {
		os.Setenv("_CM_AWS_REGION", defaultRegion)
	} else {
		log.Fatal("env AWS_REGION or AWS_DEFAULT_REGION required")
	}

	ecssvc := svc.NewECSService()
	instances, err := ecssvc.ListClusterInstances(*cluster)
	if err != nil {
		log.Fatal(err)
	}
	targetInstance := model.ClusterInstance{}
	for _, v := range instances {
		if v.InstanceID == *instance {
			targetInstance = v
		}
	}
	if targetInstance.InstanceID == "" {
		log.Fatalf("not found instanceID: %s", *instance)
	}
	if err := ecssvc.DrainContainerInstance(targetInstance); err != nil {
		log.Fatal(err)
	}
	log.Println("Draining instance: ", *instance)

	start := time.Now()
	for {
		time.Sleep(10 * time.Second)

		taskCount, err := ecssvc.CountTasksContainerInstance(targetInstance)
		if err != nil {
			log.Fatal(err)
		}
		diff := time.Since(start)
		if taskCount == 0 {
			log.Println("Tasks disappeared on instance ", *instance, ". Will deregister and terminate.")
			break
		} else {
			log.Println("#Tasks running on instance: ", taskCount)
		}

		if int64(diff.Seconds()) > *maxDrainWait {
			log.Println("Timeout wait for draining")
			break
		}
	}

	if err := ecssvc.DeregisterContainerInstance(targetInstance); err != nil {
		log.Fatal(err)
	}
	log.Println("Deregistered instance: ", *instance)
	ec2svc := svc.NewEC2Service()
	if err := ec2svc.DestroyInstance(targetInstance); err != nil {
		log.Fatal(err)
	}
	log.Println("Terminated instance: ", *instance)
}
