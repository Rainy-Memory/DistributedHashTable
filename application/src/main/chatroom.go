package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	yellow = NewColor("\033[1;33m")
	cyan   = NewColor("\033[0;36m")
	green  = NewColor("\033[0;32m")
	red    = NewColor("\033[0;31m")
)

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

func RunChatRoomCommandLine() {
	yellow.Println("Welcome to dht chat room!")
	cyan.Println("Before using it, you are required to take a few steps to register.")
	cyan.Println("Please input a port to continue...")
	green.Println("If you have no idea what port should be chosen, just randomly pick one from 20000 to 20100.")
	var port int
	_, _ = fmt.Scanf("%d\n", &port)
	node := NewNode(port)
	go node.Run()
	green.Println("DHT Node initialize finished.")
	cyan.Println("Now choose to create a new dht system or joined an existing one.")
	cyan.Println("If your friend has created one, you can join his/her system; otherwise create one for yourself.")
	cyan.Println("Please input [create] or [join] to select.")
	choice := GetArgInScope([]string{"create", "join"})
	if choice == "create" {
		node.Create()
		green.Println("Successfully create a new dht system.")
	} else {
		joinFlag := true
		for joinFlag {
			cyan.Println("Input the ip address you want to join.")
			var addr string
			_, _ = fmt.Scanf("%s\n", &addr)
			ok := node.Join(addr)
			if ok {
				green.Println("Successfully join an existing dht system.")
				joinFlag = false
			} else {
				red.Println("Failed to join this dht system. Please try again.")
				node.Quit()
			}
		}
	}
	cyan.Println("Now you need a username.")
	red.Println("NOTICE THAT USERNAME CAN'T BE CHANGED.")
	name := GetWithout([]string{"system", "System", "SYSTEM"}, "Username is same with a reserved words.")
	cyan.Println("It's all done. Now you can continue with our chatroom application.")
	green.Println("Input \"help\" to get more information.")
	flag := true
	for flag {
		text := GetLine()
		args := strings.Fields(text)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "help":
			yellow.Println("Now supported command:")
			yellow.Println("-------------------------------------------------------")
			yellow.Println("[help]              Print help message.")
			yellow.Println("[address]           Get your current address.")
			yellow.Println("[new_room <name>]   Create a new chatroom named <name>.")
			yellow.Println("[enter <name>]      Enter chatroom named <name>.")
			yellow.Println("[exit]              Exit application.")
			yellow.Println("-------------------------------------------------------")
			cyan.Println("**NOTICE** When you are in a chatroom, a new command line rules is used.")
			cyan.Println("**NOTICE** Input help in a chatroom to get more information.")
		case "address":
			if len(args) != 1 {
				red.Println("Wrong argument number!")
				break
			}
			cyan.Printf("Your current address is [%v].\n", node.GetAddress())
		case "new_room":
			if len(args) != 2 {
				red.Println("Wrong argument number!")
				break
			}
			roomName := args[1]
			have, _ := node.GetMessageNumber(roomName)
			if have {
				red.Println("This room is already exist! You can enter it directly.")
				break
			}
			ok := node.NewRoom(roomName)
			if !ok {
				red.Println("Create chatroom failed.")
				break
			}
			roomPrivateKey, roomPublicKey := GenerateRSAKey(RSABits)
			node.PutRoomPublicKey(roomName, roomPublicKey)
			go func() {
				num := 0
				for {
					newNum := node.GetUserNumber(roomName)
					if newNum > num {
						for i := num + 1; i <= newNum; i++ {
							userPublicKey := node.GetUserPublicKey(roomName, i)
							node.DeleteUserPublicKey(roomName, i)
							for j := 0; j <= PackNumber; j++ {
								var part []byte
								if j == PackNumber {
									part = roomPrivateKey[j*MaxEncryptedSize:]
								} else {
									part = roomPrivateKey[j*MaxEncryptedSize : (j+1)*MaxEncryptedSize]
								}
								node.PutUserEncryptedPart(roomName, i, j, Encrypt(part, userPublicKey))
							}
						}
						num = newNum
					}
					time.Sleep(50 * time.Millisecond)
				}
			}()
			green.Println("Successfully created chatroom.")
		case "enter":
			if len(args) != 2 {
				red.Println("Wrong argument number!")
				break
			}
			roomName := args[1]
			quitFlag := false

			// get message number
			ok, messageNumber := node.GetMessageNumber(roomName)
			if !ok {
				red.Println("Wrong room name!")
				break
			}

			// get room private key
			userPrivateKey, userPublicKey := GenerateRSAKey(RSABits)
			userNumber := node.GetUserNumber(roomName)
			userNumber++
			node.PutUserPublicKey(roomName, userNumber, userPublicKey)
			node.PutUserNumber(roomName, userNumber)
			time.Sleep(200 * time.Millisecond)
			ciphertext := node.GetUserEncryptedPart(roomName, userNumber, 0)
			roomPrivateKey := Decrypt(ciphertext, userPrivateKey)
			for j := 1; j <= PackNumber; j++ {
				ciphertext = node.GetUserEncryptedPart(roomName, userNumber, j)
				roomPrivateKey = append(roomPrivateKey, Decrypt(ciphertext, userPrivateKey)...)
			}
			roomPublicKey := node.GetRoomPublicKey(roomName)
			node.DeleteUserEncrypted(roomName, userNumber, PackNumber)

			// enter message
			enterMsg := fmt.Sprintf("[system] user [%v] entered this chat room.", name)
			messageNumber++
			enterMsg = bytes2str(Encrypt(str2bytes(enterMsg), roomPublicKey))
			node.PutMessage(roomName, messageNumber, enterMsg)
			node.PutMessageNumber(roomName, messageNumber)

			var flagLock sync.RWMutex
			green.Printf("Welcome to chatroom [%v]! Input \">help\" to get more information.\n", roomName)

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
					innerOk, msg := node.GetMessage(roomName, num+1)
					if innerOk {
						num++
						msg = bytes2str(Decrypt(str2bytes(msg), roomPrivateKey))
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
					yellow.Println("Now supported command (in chatroom):")
					yellow.Println("----------------------------------------")
					yellow.Println("[>help]      Print help message.")
					yellow.Println("[>history]   Show chat history.")
					yellow.Println("[>leave]     Left this chatroom.")
					yellow.Println("default      Send this line to chatroom.")
					yellow.Println("----------------------------------------")
				case ">history":
					cyan.Println("Chat history:")
					cyan.Println("---------------------------------------------")
					_, messageNumber = node.GetMessageNumber(roomName)
					for i := 1; i <= messageNumber; i++ {
						_, innerMsg := node.GetMessage(roomName, i)
						innerMsg = bytes2str(Decrypt(str2bytes(innerMsg), roomPrivateKey))
						fmt.Println(innerMsg)
					}
					cyan.Println("---------------------------------------------")
				case ">leave":
					flagLock.Lock()
					quitFlag = true
					flagLock.Unlock()
					green.Printf("Successfully leave room [%v].\n", roomName)
					leaveMsg := fmt.Sprintf("[system] user [%v] left this chat room.\n", name)
					_, messageNumber = node.GetMessageNumber(roomName)
					messageNumber++
					leaveMsg = bytes2str(Encrypt(str2bytes(leaveMsg), roomPublicKey))
					node.PutMessage(roomName, messageNumber, leaveMsg)
					node.PutMessageNumber(roomName, messageNumber)
				default:
					if len(str2bytes(msg)) > MaxMessageLength {
						red.Println("Message exceeded max size: 90.")
						break
					}
					msg = fmt.Sprintf("[%v][%v] %v", time.Now().Format("2006-01-02 15:04:05"), name, msg)
					_, messageNumber = node.GetMessageNumber(roomName)
					messageNumber++
					msg = bytes2str(Encrypt(str2bytes(msg), roomPublicKey))
					node.PutMessage(roomName, messageNumber, msg)
					node.PutMessageNumber(roomName, messageNumber)
				}
			}
		case "exit":
			if len(args) != 1 {
				red.Println("Wrong argument number!")
				break
			}
			flag = false
			green.Println("Successfully exit application. Have a nice day!")
		default:
			red.Println("Wrong command type!")
		}
	}
	node.Quit()
}
