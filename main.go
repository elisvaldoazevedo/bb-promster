package main

import (
	"text/template"
	"fmt"
	"os"
	"time"
	"go.etcd.io/etcd/clientv3"
	"context"
	"path"
)

type Version struct{
	PilotVersion string
	ProdVersion string
}

func main() {
	
	//etcd writing
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://etcd:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil { panic(err)}

	
	_, err = cli.Put(context.TODO(), "/pilot_version/v0003", "")
	if err != nil {
    	fmt.Print(err)
	}

	_, err = cli.Put(context.TODO(), "/prod_version/v0001", "")
	if err != nil {
    	fmt.Print(err)
	}
	
	
	//getting data from etcd
	kv, err := cli.Get(context.TODO(), "/pilot_version", clientv3.WithPrefix())
	pilot_version := path.Base(string(kv.Kvs[0].Key))

	kv2, err := cli.Get(context.TODO(), "/prod_version", clientv3.WithPrefix())
	prod_version := path.Base(string(kv2.Kvs[0].Key))

	
	
	
	if err != nil {
    	fmt.Print(err)
	}

	defer cli.Close()
	
	versionsChan := make(chan []string, 0)
	go watchUpdatedVersions("/prod_version", cli, versionsChan)
	generateAlertFile(pilot_version, prod_version)
	fmt.Print(versionsChan)
}

func generateAlertFile(pilot_version string, prod_version string) {
	fmt.Print("Generate file")
	// templating 
	versions := Version{pilot_version, prod_version}
	tmpl, err := template.ParseFiles("/etc/prometheus/alert-rules.yml.tmpl")
	if err != nil { panic(err) }

	// file creating
	f, err := os.Create("/etc/prometheus/comparative-alerts.yml")
	if err != nil {
		fmt.Println("create file: ", err)
		return
	}
	err = tmpl.Execute(f, versions)
	if err != nil { panic(err) }
}

//it works, but isn't being called by go thread
func watchUpdatedVersions(version string, cli *clientv3.Client, versionsChan chan []string){
	fmt.Print("function called")
	rch := cli.Watch(context.Background(), version)
	for wresp := range rch {
    	for _, ev := range wresp.Events {
        	fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
    	}
	}
}