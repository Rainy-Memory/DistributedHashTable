package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
)

func mytest() {
	log.SetOutput(os.Stdout)
	var level string
	fmt.Printf("Input log level\n")
	_, _ = fmt.Scanf("%s", &level)
	switch level {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.ErrorLevel)
	}
	CommandLine()
}

const (
	MaxNodeSize = 1000
)

func removeNodeFromArray(s []int, num int) (ret []int, ok bool) {
	ret = make([]int, 0, MaxNodeSize)
	ok = false
	for i := 0; i < len(s); i++ {
		if s[i] != num {
			ret = append(ret, s[i])
		} else {
			ok = true
		}
	}
	return
}

func CommandLine() {
	var cmdType, arg1, arg2 string
	flag := true
	nodeCnt := 0
	port := 20000
	localAddress := GetLocalAddress()
	nodes := new([MaxNodeSize]dhtNode)
	nodeAddresses := new([MaxNodeSize]string)
	nodesInNetwork := make([]int, 0, MaxNodeSize)
	kvMap := make(map[string]string)

	fmt.Println("Welcome to dht command line debug system!")
	fmt.Println("Input \"help\" to get more information.")
	for flag {
		_, _ = fmt.Scanf("%s %s %s", &cmdType, &arg1, &arg2)
		// fmt.Printf("type[%v], key[%v], value[%v]\n", cmdType, k, v)
		switch cmdType {
		case "help":
			fmt.Println("Now supported command:")
			fmt.Println("--------------------------------------------------------------------------------")
			fmt.Println("[add]                  Add a new node to dht system.")
			fmt.Println("[random_put <number>]  Put k-v pair [1~<number>][random string] into dictionary.")
			fmt.Println("[put <key> <value>]    Put k-v pair [<key>][<value>] into dictionary.")
			fmt.Println("[get <key>]            Get the value corresponding to <key>.")
			fmt.Println("[delete <key>]         Delete the k-v pair corresponding to <key> in dictionary.")
			fmt.Println("[quit <n>]             Quit node <n> from dht system.")
			fmt.Println("[force_quit <n>]       Force quit node <n> from dht system.")
			fmt.Println("[print map]            Print all k-v pair stored in system.")
			fmt.Println("[print node]           Print all nodes in system.")
			fmt.Println("[check]                Check whether dht is same with std.")
			fmt.Println("[exit]                 Exit CommandLine system.")
			fmt.Println("--------------------------------------------------------------------------------")
		case "add":
			nodes[nodeCnt] = NewNode(port)
			nodeAddresses[nodeCnt] = localAddress + ":" + strconv.Itoa(port)
			go nodes[nodeCnt].Run()
			success := false
			if nodeCnt == 0 {
				nodes[nodeCnt].Create()
				success = true
			} else {
				success = nodes[nodeCnt].Join(nodeAddresses[nodesInNetwork[rand.Intn(len(nodesInNetwork))]])
			}
			nodesInNetwork = append(nodesInNetwork, nodeCnt)
			nodeCnt++
			port++
			if success {
				fmt.Printf("Add node No.%v successfully.\n", nodeCnt)
			} else {
				fmt.Printf("Add node No.%v failed.\n", nodeCnt)
			}
		case "random_put":
			if nodeCnt == 0 {
				fmt.Println("No nodes in system!")
			} else {
				num, _ := strconv.Atoi(arg1)
				for i := 1; i <= num; i++ {
					k, v := strconv.Itoa(i), randString(10)
					kvMap[k] = v
					ok := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Put(k, v)
					if ok {
						fmt.Printf("Put [key:%v][value:%v] successfully.\n", k, v)
					} else {
						fmt.Printf("Put [key:%v][value:%v] failed.\n", k, v)
					}
				}
			}
		case "put":
			if nodeCnt == 0 {
				fmt.Println("No nodes in system!")
			} else {
				kvMap[arg1] = arg2
				ok := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Put(arg1, arg2)
				if ok {
					fmt.Println("Put successfully.")
				} else {
					fmt.Println("Put failed.")
				}
			}
		case "get":
			if nodeCnt == 0 {
				fmt.Println("No nodes in system!")
			} else {
				vMap, okMap := kvMap[arg1]
				ok, v := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Get(arg1)
				if ok == okMap && v == vMap {
					fmt.Println("Get successfully.")
				} else {
					fmt.Println("Get failed.")
				}
				fmt.Printf("value:[%v], ok:[%v], std value:[%v], std ok:[%v]\n", v, ok, vMap, okMap)
			}
		case "delete":
			if nodeCnt == 0 {
				fmt.Println("No nodes in system!")
			} else {
				_, okMap := kvMap[arg1]
				delete(kvMap, arg1)
				ok := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Delete(arg1)
				if ok == okMap {
					fmt.Println("Delete successfully.")
				} else {
					fmt.Println("Delete failed.")
				}
			}
		case "quit":
			num, _ := strconv.Atoi(arg1)
			num--
			nodes[num].Quit()
			var ok bool
			nodesInNetwork, ok = removeNodeFromArray(nodesInNetwork, num)
			if !ok {
				fmt.Println("Node serial number error!")
			} else {
				fmt.Println("Quit node successfully.")
			}
		case "force_quit":
			num, _ := strconv.Atoi(arg1)
			num--
			nodes[num].ForceQuit()
			var ok bool
			nodesInNetwork, ok = removeNodeFromArray(nodesInNetwork, num)
			if !ok {
				fmt.Println("Node serial number error!")
			} else {
				fmt.Println("Force quit node successfully.")
			}
		case "print":
			switch arg1 {
			case "map":
				for k, v := range kvMap {
					fmt.Printf("key [%v], value [%v]\n", k, v)
				}
			case "node":
				for _, val := range nodesInNetwork {
					fmt.Print(val+1, " ")
				}
				fmt.Println()
			}
		case "check":
			if nodeCnt == 0 {
				fmt.Println("No nodes in system!")
			} else {
				checkFlag := false
				for kMap, vMap := range kvMap {
					ok, v := nodes[nodesInNetwork[rand.Intn(len(nodesInNetwork))]].Get(kMap)
					if !ok {
						fmt.Printf("Get key %v failed.\n", kMap)
						checkFlag = true
					} else if v != vMap {
						fmt.Printf("k-v pair is not synchronized with std, expected %v, get %v.", vMap, v)
						checkFlag = true
					}
				}
				if !checkFlag {
					fmt.Println("DHT system is same with std.")
				}
			}
		case "exit":
			flag = false
			fmt.Println("Successfully exit CommandLine.")
		default:
			fmt.Println("Error command type!")
		}
	}
}
