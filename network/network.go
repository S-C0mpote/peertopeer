package peer

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type peer struct {
	conn     net.Conn
	addr     string
	nickname string
}

type Network struct {
	peers map[net.Conn]peer

	Nickname string
	Port     int
	OnReady  func()
}

func (network *Network) Listen(contact string, isFirst bool) {
	network.peers = make(map[net.Conn]peer, 0)

	if !(isFirst) {
		conn, err := net.Dial("tcp", contact)
		if err != nil {
			log.Fatal("Impossible de se connecter à l'adresse ", contact)
		}

		network.peers[conn] = peer{
			conn: conn,
			addr: contact,
		}

		network.handleArrival(conn)
		network.DisplayNetwork()

		for _, peer := range network.peers {
			go network.messageListener(peer.conn)
		}
	} else {
		fmt.Println("En attente d'une connexion entrante...")
	}

	network.OnReady()

	listenPort := fmt.Sprint(":", network.Port)
	listener, err := net.Listen("tcp", listenPort)
	if err != nil {
		log.Fatal("Impossible d'écouter à l'adresse ", listenPort)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print("Problème à l'arrivée d'un pair")
			continue
		}
		ip, nickname, ok := network.handleConnection(conn)
		if ok {
			network.peers[conn] = peer{conn: conn, addr: ip, nickname: nickname}
			fmt.Println("INFO - Connexion de " + network.peers[conn].nickname + " (" + network.peers[conn].addr + ")")
			go network.messageListener(conn)
		}
	}
}

func (network *Network) handleConnection(conn net.Conn) (string, string, bool) {
	reader := bufio.NewReader(conn)

	msg, err := reader.ReadString('\n')
	if err != nil {
		log.Print("Problème à l'arrivée d'un pair (réception de message)")
		return "", "", false
	}

	args := strings.Split(msg, ":::")

	switch args[0] {
	case "arrival":
		writer := bufio.NewWriter(conn)

		for _, peer := range network.peers {
			toSend := fmt.Sprint("peer:::", peer.addr, ":::", peer.nickname, "\n")
			writer.WriteString(toSend)
		}

		writer.WriteString("user-info:::" + network.Nickname + "\n")
		writer.WriteString("done\n")
		writer.Flush()
	}

	ip := conn.RemoteAddr().String()
	splitPoint := strings.LastIndex(ip, ":")
	ip = ip[:splitPoint]
	ip += ":" + args[1]
	nickname := strings.TrimSuffix(args[2], "\n")
	return ip, nickname, true
}

func (network *Network) handleArrival(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	toSend := fmt.Sprint("arrival:::", network.Port, ":::", network.Nickname, "\n")
	_, err1 := writer.WriteString(toSend)
	err2 := writer.Flush()
	if err1 != nil || err2 != nil {
		log.Fatal("Impossible d'écrire à mon point d'entrée, tant pis, j'arrête ici")
	}

loop:
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Print("J'ai mal lu un message, je rate peut-être un autre pair")
			continue
		}

		args := strings.Split(msg, ":::")

		switch args[0] {
		case "peer":
			fullAddr := args[1]
			nickname := strings.TrimSuffix(args[2], "\n")

			newConn := network.getInTouch(fullAddr)
			if newConn != nil {
				network.peers[newConn] = peer{
					conn:     newConn,
					addr:     fullAddr,
					nickname: nickname,
				}
			}
		case "user-info":
			pseudo := strings.TrimSuffix(args[1], "\n")
			contactPeer, found := network.peers[conn]

			if !found {
				log.Println("USERINFO: Error user not found")
			} else {
				network.peers[conn] = peer{
					conn:     conn,
					addr:     contactPeer.addr,
					nickname: pseudo,
				}
			}
		case "done\n":
			break loop
		}
	}
}

func (network *Network) messageListener(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("INFO - Déconnexion de " + network.peers[conn].nickname + " (" + network.peers[conn].addr + ")")
			delete(network.peers, conn)
			return
		}

		args := strings.Split(msg, ":::")

		switch args[0] {
		case "broadcast":
			msg := strings.TrimSuffix(args[1], "\n")
			peer, _ := network.peers[conn]
			fmt.Println("[" + peer.nickname + "] " + msg)
		default:
			log.Println("Not found: " + msg)
		}
	}
}

func (network *Network) BroadcastMessage(message string) {
	for _, peer := range network.peers {
		writer := bufio.NewWriter(peer.conn)
		writer.WriteString("broadcast:::" + message)
		writer.Flush()
	}
}

func (network *Network) SendPrivateMessage(username string, message string) {

}

func (network *Network) getInTouch(fullAddr string) net.Conn {
	log.Print("On vient de me donner l'adresse ", fullAddr, ", j'ouvre tout de suite une connexion !")
	conn, err := net.Dial("tcp", fullAddr)
	if err != nil {
		log.Print("Attention : impossible de joindre ", fullAddr)
		return nil
	}

	writer := bufio.NewWriter(conn)
	msg := fmt.Sprint("connection:::", network.Port, ":::", network.Nickname, "\n")
	_, err1 := writer.WriteString(msg)
	err2 := writer.Flush()
	if err1 != nil || err2 != nil {
		log.Print("Attention : impossible de se signaler à ", fullAddr)
	}

	return conn
}

func (network *Network) DisplayNetwork() {
	addr := ""
	for _, peer := range network.peers {
		addr += "\t" + peer.addr + " | " + peer.nickname + "\n"
	}
	fmt.Print("Liste des personnes connectées :\n", addr)
}
