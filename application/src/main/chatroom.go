package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetLineAndSeparate() []string {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error occurred! Error message: [%v].\n", err)
		return nil
	}
	return strings.Fields(text)
}

func GetLine() string {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error occurred! Error message: [%v].\n", err)
		return ""
	}
	return text
}

func GetWithout(scope []string, errMsg string) string {
	for {
		var text string
		_, _ = fmt.Scanf("%s\n", &text)
		flag := true
		for _, target := range scope {
			if text == target {
				flag = false
			}
		}
		if flag {
			return text
		} else {
			fmt.Println(errMsg)
		}
	}
}

func GetArgInScope(scope []string) string {
	for {
		var text string
		_, _ = fmt.Scanf("%s\n", &text)
		for _, target := range scope {
			if text == target {
				return text
			}
		}
		fmt.Println("Wrong arguments!")
	}
}

/*
storage format:
{[name] message number} -> chatroom [name] message number
{[name] [number]}       -> chatroom [name] No.[number] message
*/

func RunChatRoomCommandLine() {
	fmt.Println("Welcome to dht chat room!")
	fmt.Println("Before using it, you are required to take a few steps to register.")
	fmt.Println("Please input a port to continue...")
	fmt.Println("If you have no idea what port should be chosen, just randomly pick one from 20000 to 20100.")
	var port int
	_, _ = fmt.Scanf("%d\n", &port)
	node := NewNode(port)
	go node.Run()
	fmt.Println("DHT Node initialize finished.")
	fmt.Println("Now choose to create a new dht system or joined an existing one.")
	fmt.Println("If your friend has created one, you can join his/her system; otherwise create one for yourself.")
	fmt.Println("Please input [create] or [join] to select.")
	choice := GetArgInScope([]string{"create", "join"})
	if choice == "create" {
		node.Create()
		fmt.Println("Successfully create a new dht system.")
	} else {
		joinFlag := true
		for joinFlag {
			fmt.Println("Input the ip address you want to join.")
			var addr string
			_, _ = fmt.Scanf("%s\n", &addr)
			ok := node.Join(addr)
			if ok {
				fmt.Println("Successfully join an existing dht system.")
				joinFlag = false
			} else {
				fmt.Println("Failed to join this dht system. Please try again.")
				node.Quit()
			}
		}
	}
	fmt.Println("Now you need a username.")
	fmt.Println("NOTICE THAT USERNAME CAN'T BE CHANGED.")
	name := GetWithout([]string{"system", "System", "SYSTEM"}, "Username is same with a reserved words.")
	fmt.Println("It's all done. Now you can continue with our chatroom application.")
	fmt.Println("Input \"help\" to get more information.")
	flag := true
	for flag {
		args := GetLineAndSeparate()
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "help":
			fmt.Println("Now supported command:")
			fmt.Println("-------------------------------------------------------")
			fmt.Println("[help]              Print help message.")
			fmt.Println("[address]           Get your current address.")
			fmt.Println("[new_room <name>]   Create a new chatroom named <name>.")
			fmt.Println("[enter <name>]      Enter chatroom named <name>.")
			fmt.Println("[exit]              Exit application.")
			fmt.Println("-------------------------------------------------------")
			fmt.Println("**NOTICE** When you are in a chatroom, a new command line rules is used.")
			fmt.Println("**NOTICE** Input help in a chatroom to get more information.")
		case "address":
			if len(args) != 1 {
				fmt.Println("Wrong argument number!")
				break
			}
			fmt.Printf("Your current address is [%v].\n", node.GetAddress())
		case "new_room":
			if len(args) != 2 {
				fmt.Println("Wrong argument number!")
				break
			}
			have, _ := node.Get(appHash(args[1] + " message number"))
			if have {
				fmt.Println("This room is already exist! You can enter it directly.")
				break
			}
			ok := node.Put(appHash(args[1]+" message number"), "0")
			if ok {
				fmt.Println("Successfully created chatroom.")
			} else {
				fmt.Println("Create chatroom failed.")
			}
		case "enter":
			if len(args) != 2 {
				fmt.Println("Wrong argument number!")
				break
			}
			roomName := args[1]
			quitFlag := false
			// get message number
			ok, messageNumberStr := node.Get(appHash(roomName + " message number"))
			if !ok {
				fmt.Println("Wrong room name!")
				break
			}
			messageNumber, _ := strconv.Atoi(messageNumberStr)
			enterMsg := fmt.Sprintf("[system] user [%v] entered this chat room.", name)
			messageNumber++
			node.Put(appHash(roomName+" message "+strconv.Itoa(messageNumber)), enterMsg)
			node.Put(appHash(roomName+" message number"), strconv.Itoa(messageNumber))

			var flagLock sync.RWMutex
			fmt.Printf("Welcome to chatroom [%v]! Input \">help\" to get more information.\n", roomName)
			// accept message
			go func(flag *bool, number int) {
				num := number
				for {
					flagLock.RLock()
					temp := *flag
					flagLock.RUnlock()
					if temp {
						break
					}
					innerOk, msg := node.Get(appHash(roomName + " message " + strconv.Itoa(num+1)))
					if innerOk {
						num++
						fmt.Println(msg)
					}
					time.Sleep(100 * time.Millisecond)
				}
			}(&quitFlag, messageNumber)
			// inner command line
			for {
				flagLock.RLock()
				temp := quitFlag
				flagLock.RUnlock()
				if temp {
					break
				}
				msg := GetLine()
				msg = strings.TrimRight(msg, "\n")
				innerArgs := strings.Fields(msg)
				if len(innerArgs) == 0 {
					continue
				}
				switch innerArgs[0] {
				case ">help":
					fmt.Println("Now supported command (in chatroom):")
					fmt.Println("----------------------------------------")
					fmt.Println("[>help]      Print help message.")
					fmt.Println("[>history]   Show chat history.")
					fmt.Println("[>leave]     Left this chatroom.")
					fmt.Println("default      Send this line to chatroom.")
					fmt.Println("----------------------------------------")
				case ">history":
					fmt.Println("Chat history:")
					fmt.Println("---------------------------------------------")
					_, messageNumberStr = node.Get(appHash(roomName + " message number"))
					messageNumber, _ = strconv.Atoi(messageNumberStr)
					for i := 1; i <= messageNumber; i++ {
						_, innerMsg := node.Get(appHash(roomName + " message " + strconv.Itoa(i)))
						fmt.Println(innerMsg)
					}
					fmt.Println("---------------------------------------------")
				case ">leave":
					flagLock.Lock()
					quitFlag = true
					flagLock.Unlock()
					fmt.Printf("Successfully leave room [%v].\n", roomName)
					leaveMsg := fmt.Sprintf("[system] user [%v] left this chat room.\n", name)
					_, messageNumberStr = node.Get(appHash(roomName + " message number"))
					messageNumber, _ = strconv.Atoi(messageNumberStr)
					messageNumber++
					node.Put(appHash(roomName+" message "+strconv.Itoa(messageNumber)), leaveMsg)
					node.Put(appHash(roomName+" message number"), strconv.Itoa(messageNumber))
				default:
					msg = fmt.Sprintf("[%v][%v] %v", time.Now().Format("2006-01-02 15:04:05"), name, msg)
					_, messageNumberStr = node.Get(appHash(roomName + " message number"))
					messageNumber, _ = strconv.Atoi(messageNumberStr)
					messageNumber++
					node.Put(appHash(roomName+" message "+strconv.Itoa(messageNumber)), msg)
					node.Put(appHash(roomName+" message number"), strconv.Itoa(messageNumber))
				}
			}
		case "exit":
			if len(args) != 1 {
				fmt.Println("Wrong argument number!")
				break
			}
			flag = false
			fmt.Println("Successfully exit application. Have a nice day!")
		default:
			fmt.Println("Wrong command type!")
		}
	}
	node.Quit()
}
