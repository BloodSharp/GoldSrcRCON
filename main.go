package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"regexp"
)

func main() {
	rhost := flag.String("rhost", "127.0.0.1:27015", "Sets the remote server host")
	rconPassword := flag.String("rcon", "", "Sets the remote rcon password")
	rconCommand := flag.String("command", "status", "Sets the remote command to execute")
	flag.Usage = usageHelp
	flag.Parse()

	if len(os.Args) <= 1 {
		usageHelp()
		return
	}

	serverConnection, err := net.Dial("udp", *rhost)
	if err != nil {
		fmt.Errorf("Error trying to connect to %v ==> %v", *rhost, err.Error())
		return
	}
	defer serverConnection.Close()

	fmt.Println(sendRCON(serverConnection, *rconPassword, *rconCommand))
}

func usageHelp() {
	fmt.Println("BloodSharp's GoldSrc RCON Help")
	flag.PrintDefaults()
}

func prepareCommand(command string) []byte {
	var bytesSequence []byte
	//intial 4 characters as per standard
	bytesSequence = append(bytesSequence, []byte{255, 255, 255, 255}...)
	//copying bytes from challenge rcon to send buffer
	bytesSequence = append(bytesSequence, []byte(command)...)
	return bytesSequence
}

func sendRCON(serverConnection net.Conn, rconPassword string, rconCommand string) string {
	//sending challenge command to counter strike server
	bufferSend := prepareCommand("challenge rcon\n")
	_, err := serverConnection.Write(bufferSend)
	if err != nil {
		fmt.Errorf("Error trying to request challenge to %v ==> %v",
			serverConnection.RemoteAddr().String(), err.Error())
		return "[Error]"
	}

	//get challenge response
	var bufferReceive [4000]byte
	_, err = serverConnection.Read(bufferReceive[:])
	if err != nil {
		fmt.Errorf("Error trying to receive challenge from %v ==> %v",
			serverConnection.RemoteAddr().String(), err.Error())
		return "[Error]"
	}

	//retrive number from challenge response
	challengeRCON := string(bufferReceive[:])
	challengeRCONSplitted := regexp.MustCompile("[^\\d]").Split(challengeRCON, -1)[:]
	challengeRCON = ""
	for i := 0; i < len(challengeRCONSplitted); i++ {
		challengeRCON += challengeRCONSplitted[i]
	}

	//preparing rcon command to send
	commandToSend := "rcon \"" + challengeRCON + "\" " + rconPassword + " " + rconCommand + "\n"
	bufferSend = prepareCommand(commandToSend)

	_, err = serverConnection.Write(bufferSend)
	if err != nil {
		fmt.Errorf("Error trying to send the command to %v ==> %v",
			serverConnection.RemoteAddr().String(), err.Error())
		return "[Error]"
	}

	_, err = serverConnection.Read(bufferReceive[:])
	if err != nil {
		fmt.Errorf("Error trying to get the command result from %v ==> %v",
			serverConnection.RemoteAddr().String(), err.Error())
		return "[Error]"
	}

	//strip initial bytes from command response
	var stringResponse string
	if bufferReceive[0] == 255 && bufferReceive[1] == 255 &&
		bufferReceive[2] == 255 && bufferReceive[3] == 255 && bufferReceive[4] == 'l' {
		stringResponse = string(bufferReceive[5:])
	} else {
		stringResponse = string(bufferReceive[:])
	}
	return stringResponse
}
