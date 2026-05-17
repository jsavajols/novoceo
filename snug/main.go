// snug : relie un bouton Zigbee à une prise connectée via MQTT.
// À chaque pression du bouton, envoie un TOGGLE sur le topic /set de la prise.
// Supporte les topologies locale (zigbee2mqtt/<device>) et multi-RPi (zigbee2mqtt/<rpi>/<device>).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func log(format string, args ...any) {
	fmt.Printf("[snug] "+format+"\n", args...)
}

func main() {
	// broker := flag.String("broker", "tcp://localhost:1883", "URL du broker MQTT")
	broker := flag.String("broker", "tcp://192.168.1.128:32500", "URL du broker MQTT")
	button := flag.String("button", "Bouton", "Friendly name du bouton")
	plug := flag.String("plug", "", "Friendly name de la prise (requis)")
	action := flag.String("action", "", "Action bouton à filtrer (vide = toutes)")
	rpiFilter := flag.String("rpi", "", "Hostname RPi à cibler (vide = auto-détecté)")
	flag.Parse()

	if *plug == "" {
		fmt.Fprintln(os.Stderr, "Erreur : -plug est requis (ex: -plug prise_bureau)")
		flag.Usage()
		os.Exit(1)
	}

	opts := mqtt.NewClientOptions().
		AddBroker(*broker).
		SetClientID(fmt.Sprintf("snug-%d", time.Now().Unix())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			log("Connexion perdue : %v — reconnexion...", err)
		})

	var client mqtt.Client

	// Le wildcard MQTT + couvre exactement un niveau : capte zigbee2mqtt/<rpi>/<button>
	// sans intercepter les topics bridge/* ou les sous-topics imbriqués.
	topicLocal := fmt.Sprintf("zigbee2mqtt/%s", *button)
	topicRemote := fmt.Sprintf("zigbee2mqtt/+/%s", *button)

	handler := func(_ mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		rest := strings.TrimPrefix(topic, "zigbee2mqtt/")
		parts := strings.SplitN(rest, "/", 2)

		var rpi, publishTopic string
		if len(parts) == 2 {
			// format zigbee2mqtt/<rpi>/<button>
			rpi = parts[0]
			// Filtre utile quand plusieurs RPi publient le même friendly_name de bouton.
			if *rpiFilter != "" && rpi != *rpiFilter {
				return
			}
			publishTopic = fmt.Sprintf("zigbee2mqtt/%s/%s/set", rpi, *plug)
		} else {
			// format zigbee2mqtt/<button> (local)
			rpi = "local"
			publishTopic = fmt.Sprintf("zigbee2mqtt/%s/set", *plug)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			return
		}

		actionVal, _ := payload["action"].(string)
		if actionVal == "" {
			return
		}
		if *action != "" && actionVal != *action {
			return
		}

		// TOGGLE évite d'avoir à connaître l'état courant de la prise.
		token := client.Publish(publishTopic, 1, false, `{"state":"TOGGLE"}`)
		token.Wait()

		ts := time.Now().Format("15:04:05")
		log("%s [%s] %s -> action=%s -> TOGGLE %s", ts, rpi, *button, actionVal, *plug)
	}

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log("Connecté à %s", *broker)
		log("En écoute sur %s et %s", topicLocal, topicRemote)
		c.Subscribe(topicLocal, 1, handler)
		c.Subscribe(topicRemote, 1, handler)
	})

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Fprintf(os.Stderr, "Erreur de connexion : %v\n", token.Error())
		os.Exit(1)
	}
	defer client.Disconnect(250)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log("Arrêt.")
}
