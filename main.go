package main

import (
	"bufio"
	"flag"
	"os"
	"peer/network"
	"strings"
)

func main() {
	port := flag.Int(
		"port", 25000,
		"Port sur lequel écouter")
	isFirst := flag.Bool(
		"first", false,
		"Pour déclencher la création d'un nouveau réseau")
	contact := flag.String(
		"contact", "127.0.0.1:25000",
		"Adresse et port de contact quand on ne crée pas un nouveau réseau")
	nickname := flag.String(
		"nickname", "Guest",
		"Pseudo de votre compte pour votre chat")

	flag.Parse()

	network := peer.Network{Port: *port, Nickname: *nickname}

	network.OnReady = func() {
		go console(&network)
	}

	network.Listen(*contact, *isFirst)
}

func console(network *peer.Network) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// TODO: Afficher toutes les commandes au lancement du programme
		msg, _ := reader.ReadString('\n')

		args := strings.Split(msg, " ")

		if strings.HasPrefix(msg, "/list") {
			network.DisplayNetwork()
		} else if strings.HasPrefix(msg, "/mp") && len(args) >= 2 {
			network.SendPrivateMessage(args[1], msg) // TODO: Retirer le pseudo du message
		} else {
			network.BroadcastMessage(msg)
		}
	}
}
