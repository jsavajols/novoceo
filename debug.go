// Outil de diagnostic : affiche sur stderr tous les messages MQTT reçus sur zigbee2mqtt/#
// pendant 60 secondes. Utile pour inspecter la structure brute des topics avant de modifier
// la logique de routage dans main.go.
package main

import (
    "fmt"
    "os"
    "time"
    mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
    opts := mqtt.NewClientOptions().
        AddBroker("tcp://localhost:1883").
        SetClientID("debug-monitor").
        SetOnConnectHandler(func(c mqtt.Client) {
            c.Subscribe("zigbee2mqtt/#", 0, func(_ mqtt.Client, msg mqtt.Message) {
                fmt.Fprintf(os.Stderr, "TOPIC: %s\nPAYLOAD: %s\n\n", msg.Topic(), msg.Payload())
            })
        })
    client := mqtt.NewClient(opts)
    client.Connect()
    time.Sleep(60 * time.Second)
}
