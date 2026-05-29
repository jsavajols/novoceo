// send-device : envoie une commande MQTT à un appareil Zigbee2MQTT depuis la ligne de commande.
// Contrairement à snug qui écoute un bouton et réagit, send-device agit directement au lancement :
// utile pour les scripts, les tests ou l'automatisation via cron/n8n/etc.
// Supporte les topologies locale (zigbee2mqtt/<device>/set) et multi-RPi (zigbee2mqtt/<rpi>/<device>/set).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func log(format string, args ...any) {
	fmt.Printf("[send-device] "+format+"\n", args...)
}

func main() {
	// broker  := flag.String("broker",  "tcp://localhost:1883", "URL du broker MQTT")
	broker := flag.String("broker", "tcp://192.168.1.128:32500", "URL du broker MQTT")
	device := flag.String("device", "", "Friendly name de l'appareil cible (requis)")
	rpi := flag.String("rpi", "", "Hostname RPi pour topologie multi-RPi (vide = local)")
	state := flag.String("state", "TOGGLE", "État à envoyer : ON, OFF, TOGGLE")
	payload := flag.String("payload", "", "Payload JSON brut (prioritaire sur -state, ex: '{\"brightness\":128}')")
	flag.Parse()

	if *device == "" {
		fmt.Fprintln(os.Stderr, "Erreur : -device est requis (ex: -device prise_bureau)")
		flag.Usage()
		os.Exit(1)
	}

	// Construit le topic de destination selon la topologie.
	// Avec -rpi : zigbee2mqtt/<rpi>/<device>/set (multi-RPi)
	// Sans -rpi : zigbee2mqtt/<device>/set        (local)
	var publishTopic string
	if *rpi != "" {
		publishTopic = fmt.Sprintf("zigbee2mqtt/%s/%s/set", *rpi, *device)
	} else {
		publishTopic = fmt.Sprintf("zigbee2mqtt/%s/set", *device)
	}

	// Construit le message : payload JSON brut si fourni, sinon enveloppe le -state.
	var msg string
	if *payload != "" {
		// Valide le JSON avant d'envoyer pour éviter de corrompre l'appareil.
		var check map[string]interface{}
		if err := json.Unmarshal([]byte(*payload), &check); err != nil {
			fmt.Fprintf(os.Stderr, "Erreur : -payload n'est pas un JSON valide : %v\n", err)
			os.Exit(1)
		}
		msg = *payload
	} else {
		msg = fmt.Sprintf(`{"state":"%s"}`, *state)
	}

	// Connexion one-shot : pas de reconnexion automatique, on échoue vite si le broker est absent.
	opts := mqtt.NewClientOptions().
		AddBroker(*broker).
		SetClientID(fmt.Sprintf("send-device-%d", time.Now().Unix())).
		SetConnectRetry(false)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Fprintf(os.Stderr, "Erreur de connexion au broker %s : %v\n", *broker, token.Error())
		os.Exit(1)
	}
	defer client.Disconnect(250)

	// QoS 1 : garantit la livraison au moins une fois (au contraire du QoS 0, fire-and-forget).
	token := client.Publish(publishTopic, 1, false, msg)
	token.Wait()
	if err := token.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Erreur de publication : %v\n", err)
		os.Exit(1)
	}

	ts := time.Now().Format("15:04:05")
	log("%s -> %s : %s", ts, publishTopic, msg)
}
