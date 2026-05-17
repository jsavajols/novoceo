package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// ----------------------------------------------------------------------------
// Modèles de données
// ----------------------------------------------------------------------------

type Device struct {
	IEEEAddress    string `json:"ieee_address"`
	FriendlyName   string `json:"friendly_name"`
	Type           string `json:"type"`
	Manufacturer   string `json:"manufacturer"`
	ModelID        string `json:"model_id"`
	NetworkAddress int    `json:"network_address"`
	Supported      bool   `json:"supported"`
	LastSeen       *int64 `json:"last_seen"`
}

type ValueUpdate struct {
	DeviceName string
	RPiName    string
	Values     map[string]interface{}
	ReceivedAt time.Time
}

// ----------------------------------------------------------------------------
// State - état partagé thread-safe
// ----------------------------------------------------------------------------

type State struct {
	mu      sync.RWMutex
	devices map[string][]Device     // clé = hostname RPi
	updates map[string]*ValueUpdate // clé = "hostname/device_name"
	events  []string
}

func newState() *State {
	return &State{
		devices: make(map[string][]Device),
		updates: make(map[string]*ValueUpdate),
	}
}

func (s *State) setDevices(rpi string, devices []Device) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var filtered []Device
	for _, d := range devices {
		// Le Coordinator est la passerelle Zigbee2MQTT elle-même, pas un capteur à afficher.
		if d.Type != "Coordinator" {
			filtered = append(filtered, d)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].FriendlyName < filtered[j].FriendlyName
	})
	s.devices[rpi] = filtered
}

func (s *State) updateValues(rpi, deviceName string, payload map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := rpi + "/" + deviceName
	prev := s.updates[key]
	update := &ValueUpdate{
		DeviceName: deviceName,
		RPiName:    rpi,
		Values:     payload,
		ReceivedAt: time.Now(),
	}
	s.updates[key] = update

	if prev != nil {
		for k, newVal := range payload {
			if oldVal, exists := prev.Values[k]; exists {
				if fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal) {
					s.addEvent(fmt.Sprintf("[%s] %s : %s %v -> %v", rpi, deviceName, k, oldVal, newVal))
				}
			} else {
				s.addEvent(fmt.Sprintf("[%s] %s : %s = %v (nouveau)", rpi, deviceName, k, newVal))
			}
		}
	} else {
		s.addEvent(fmt.Sprintf("[%s] %s : première valeur reçue", rpi, deviceName))
	}
}

func (s *State) addEvent(msg string) {
	ts := time.Now().Format("15:04:05")
	entry := fmt.Sprintf("[%s] %s", ts, msg)
	s.events = append(s.events, entry)
	// Anneau circulaire : on ne garde que les 12 derniers pour ne pas dépasser la hauteur du terminal.
	if len(s.events) > 12 {
		s.events = s.events[len(s.events)-12:]
	}
}

func (s *State) totalDevices() int {
	total := 0
	for _, devs := range s.devices {
		total += len(devs)
	}
	return total
}

// ----------------------------------------------------------------------------
// Rendu terminal
// ----------------------------------------------------------------------------

const (
	reset       = "\033[0m"
	bold        = "\033[1m"
	dim         = "\033[2m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorPurple = "\033[35m"
)

func clearScreen()           { fmt.Print("\033[2J\033[H") }
func hideCursor()            { fmt.Print("\033[?25l") }
func showCursor()            { fmt.Print("\033[?25h") }
func separator(w int) string { return strings.Repeat("─", w) }

func truncate(s string, max int) string {
	if len(s) <= max {
		return s + strings.Repeat(" ", max-len(s))
	}
	return s[:max-1] + "…"
}

func formatValue(key string, val interface{}) string {
	k := strings.ToLower(key)
	v := fmt.Sprintf("%v", val)
	switch {
	case k == "temperature":
		return fmt.Sprintf("%.1f°C", toFloat(val))
	case k == "humidity":
		return fmt.Sprintf("%.1f%%", toFloat(val))
	case k == "pressure":
		return fmt.Sprintf("%.0f hPa", toFloat(val))
	case k == "battery":
		pct := toFloat(val)
		return fmt.Sprintf("%s %.0f%%", batteryBar(pct), pct)
	case k == "linkquality":
		lqi := toFloat(val)
		return fmt.Sprintf("%s %.0f/255", lqiBar(lqi), lqi)
	case k == "occupancy":
		if val == true || v == "true" {
			return colorRed + "● DÉTECTÉ" + reset
		}
		return dim + "○ vide" + reset
	case k == "contact":
		if val == false || v == "false" {
			return colorRed + "◈ OUVERT" + reset
		}
		return colorGreen + "◉ fermé" + reset
	case k == "state":
		if strings.EqualFold(v, "on") {
			return colorGreen + "◉ ON" + reset
		}
		return dim + "○ off" + reset
	case k == "action":
		if v != "" && v != "<nil>" {
			return colorYellow + bold + "⚡ " + v + reset
		}
		return dim + "—" + reset
	case k == "voltage":
		return fmt.Sprintf("%.0f mV", toFloat(val))
	case k == "illuminance":
		return fmt.Sprintf("%.0f lx", toFloat(val))
	case k == "co2":
		return fmt.Sprintf("%.0f ppm", toFloat(val))
	default:
		return v
	}
}

// toFloat gère json.Number car le décodeur utilise UseNumber() pour éviter
// la perte de précision sur les flottants (ex : température 21.5 -> 21.499...).
func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case json.Number:
		f, _ := n.Float64()
		return f
	}
	return 0
}

func batteryBar(pct float64) string {
	bars := int(pct / 20)
	if bars > 5 {
		bars = 5
	}
	color := colorGreen
	if pct < 20 {
		color = colorRed
	} else if pct < 40 {
		color = colorYellow
	}
	return color + "[" + strings.Repeat("█", bars) + strings.Repeat("░", 5-bars) + "]" + reset
}

func lqiBar(lqi float64) string {
	stars := int(lqi / 51)
	if stars > 5 {
		stars = 5
	}
	return colorCyan + strings.Repeat("▪", stars) + strings.Repeat("·", 5-stars) + reset
}

func deviceTypeIcon(t string) string {
	switch strings.ToLower(t) {
	case "router":
		return "⟳"
	case "enddevice":
		return "◉"
	default:
		return "○"
	}
}

// ----------------------------------------------------------------------------
// Affichage principal
// ----------------------------------------------------------------------------

func render(s *State, broker string, connected bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clearScreen()
	hideCursor()

	width := 72

	// HEADER
	fmt.Printf("%s%s", bold, colorBlue)
	fmt.Printf(" ███████╗██╗ ██████╗ ██████╗ ███████╗███████╗\n")
	fmt.Printf(" ╚══███╔╝██║██╔════╝ ██╔══██╗██╔════╝██╔════╝\n")
	fmt.Printf("   ███╔╝ ██║██║  ███╗██████╔╝█████╗  █████╗  \n")
	fmt.Printf("  ███╔╝  ██║██║   ██║██╔══██╗██╔══╝  ██╔══╝  \n")
	fmt.Printf(" ███████╗██║╚██████╔╝██████╔╝███████╗███████╗\n")
	fmt.Printf(" ╚══════╝╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝\n")
	fmt.Print(reset)
	fmt.Printf(" %sMonitor%s — %s%s%s\n",
		dim, reset, dim, time.Now().Format("Monday 02 Jan 2006  15:04:05"), reset)

	statusColor, statusText, statusIcon := colorGreen, "connecté", "●"
	if !connected {
		statusColor, statusText, statusIcon = colorRed, "déconnecté", "○"
	}
	fmt.Printf(" Broker MQTT : %s%s %s%s   %s%s RPi  %s devices%s\n\n",
		statusColor, statusIcon, statusText, reset,
		colorCyan, fmt.Sprintf("%d", len(s.devices)),
		fmt.Sprintf("%d", s.totalDevices()), reset)

	// --- DEVICES par RPi ---
	fmt.Printf(" %s%s DEVICES %s\n", bold, colorCyan, reset)
	fmt.Printf(" %s%s%s\n", dim, separator(width), reset)

	// Trier les RPi par nom
	rpiNames := make([]string, 0, len(s.devices))
	for rpi := range s.devices {
		rpiNames = append(rpiNames, rpi)
	}
	sort.Strings(rpiNames)

	if len(rpiNames) == 0 {
		fmt.Printf(" %s Aucun RPi connecté — vérifie que Zigbee2MQTT tourne.%s\n\n", colorYellow, reset)
	} else {
		for _, rpi := range rpiNames {
			devs := s.devices[rpi]

			// En-tête du RPi
			fmt.Printf(" %s%s %-20s%s %s(%d device(s))%s\n",
				bold+colorPurple, "⬡", rpi, reset,
				dim, len(devs), reset)
			fmt.Printf(" %s%s%s\n", dim, separator(width), reset)

			if len(devs) == 0 {
				fmt.Printf("   %sAucun device — allume tes devices Zigbee.%s\n", dim, reset)
			} else {
				fmt.Printf("   %s%-22s %-10s %-18s%s\n",
					bold+dim, "Nom", "Type", "Modèle", reset)

				for _, d := range devs {
					modelStr := d.ModelID
					if modelStr == "" {
						modelStr = "—"
					}

					hasData := dim + " ·" + reset
					key := rpi + "/" + d.FriendlyName
					if upd, ok := s.updates[key]; ok {
						since := time.Since(upd.ReceivedAt)
						if since < 5*time.Second {
							hasData = colorGreen + " ●" + reset
						} else if since < 60*time.Second {
							hasData = colorYellow + " ○" + reset
						}
					}

					fmt.Printf("   %s%-22s%s %-10s %-18s%s\n",
						colorCyan, truncate(d.FriendlyName, 22), reset,
						deviceTypeIcon(d.Type)+" "+d.Type,
						truncate(modelStr, 18),
						hasData)
				}
			}
			fmt.Println()
		}
	}

	// --- VALEURS EN DIRECT ---
	fmt.Printf(" %s%s VALEURS EN DIRECT %s\n", bold, colorYellow, reset)
	fmt.Printf(" %s%s%s\n", dim, separator(width), reset)

	if len(s.updates) == 0 {
		fmt.Printf(" %s En attente de données...%s\n\n", dim, reset)
	} else {
		// Grouper par RPi
		byRpi := make(map[string][]*ValueUpdate)
		for _, upd := range s.updates {
			byRpi[upd.RPiName] = append(byRpi[upd.RPiName], upd)
		}

		rpiSorted := make([]string, 0, len(byRpi))
		for rpi := range byRpi {
			rpiSorted = append(rpiSorted, rpi)
		}
		sort.Strings(rpiSorted)

		for _, rpi := range rpiSorted {
			fmt.Printf(" %s%s %s%s\n", bold+colorPurple, "⬡", rpi, reset)

			updates := byRpi[rpi]
			sort.Slice(updates, func(i, j int) bool {
				return updates[i].DeviceName < updates[j].DeviceName
			})

			for _, upd := range updates {
				since := time.Since(upd.ReceivedAt).Round(time.Second)
				fmt.Printf("   %s%s%s %s(il y a %s)%s\n",
					colorGreen+bold, truncate(upd.DeviceName, 30), reset,
					dim, since, reset)

				keys := make([]string, 0, len(upd.Values))
				for k := range upd.Values {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					fmt.Printf("     %-18s %s\n", k, formatValue(k, upd.Values[k]))
				}
				fmt.Println()
			}
		}
	}

	// --- ÉVÉNEMENTS ---
	fmt.Printf(" %s%s ÉVÉNEMENTS RÉCENTS %s\n", bold, colorBlue, reset)
	fmt.Printf(" %s%s%s\n", dim, separator(width), reset)
	if len(s.events) == 0 {
		fmt.Printf(" %s Aucun événement.%s\n", dim, reset)
	} else {
		for _, ev := range s.events {
			fmt.Printf(" %s%s%s\n", dim, ev, reset)
		}
	}
	fmt.Printf("\n %s[Ctrl+C pour quitter]%s\n", dim, reset)
}

// ----------------------------------------------------------------------------
// Extraction du reste du topic MQTT après le préfixe "zigbee2mqtt/"
// Supporte le format standard : zigbee2mqtt/<device> ou zigbee2mqtt/bridge/...
// ----------------------------------------------------------------------------

func parseTopic(topic string) (rest string, ok bool) {
	rest = strings.TrimPrefix(topic, "zigbee2mqtt/")
	if rest == topic {
		return "", false
	}
	return rest, true
}

// ----------------------------------------------------------------------------
// Main
// ----------------------------------------------------------------------------

func main() {
	broker := "tcp://localhost:1883"
	if len(os.Args) > 1 {
		broker = os.Args[1]
	}

	state := newState()
	connected := false

	fmt.Printf("Connexion à %s...\n", broker)

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("zigbee-monitor-%d", time.Now().Unix())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			connected = true

			// Souscription unique - gère bridge/devices et valeurs en un seul handler
			c.Subscribe("zigbee2mqtt/#", 0, func(_ mqtt.Client, msg mqtt.Message) {
				topic := msg.Topic()
				rest, ok := parseTopic(topic)
				if !ok {
					return
				}

				// Les topics peuvent être :
				//   bridge/devices              -> liste devices locale
				//   bridge/<autre>              -> ignoré
				//   <rpi>/bridge/devices        -> liste devices du RPi
				//   <rpi>/bridge/<autre>        -> ignoré
				//   <device>                    -> valeur device local
				//   <rpi>/<device>              -> valeur device du RPi
				parts := strings.SplitN(rest, "/", 3)

				switch {
				case parts[0] == "bridge":
					if len(parts) == 2 && parts[1] == "devices" {
						var devices []Device
						if err := json.Unmarshal(msg.Payload(), &devices); err != nil {
							return
						}
						state.setDevices("local", devices)
					}
					return

				case len(parts) >= 2 && parts[1] == "bridge":
					rpi := parts[0]
					if len(parts) == 3 && parts[2] == "devices" {
						var devices []Device
						if err := json.Unmarshal(msg.Payload(), &devices); err != nil {
							return
						}
						state.setDevices(rpi, devices)
					}
					return
				}

				// Valeurs d'un device
				var rpi, deviceName string
				if len(parts) >= 2 {
					rpi = parts[0]
					deviceName = strings.Join(parts[1:], "/")
				} else {
					rpi = "local"
					deviceName = parts[0]
				}

				var payload map[string]interface{}
				dec := json.NewDecoder(strings.NewReader(string(msg.Payload())))
				dec.UseNumber() // conserve la précision numérique des capteurs (ex: 21.5°C)
				if err := dec.Decode(&payload); err != nil {
					return
				}
				state.updateValues(rpi, deviceName, payload)
			})
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			connected = false
		})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Fprintf(os.Stderr, "Erreur de connexion : %v\n", token.Error())
		os.Exit(1)
	}
	defer func() {
		showCursor()
		client.Disconnect(250)
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	hideCursor()
	for {
		select {
		case <-ticker.C:
			render(state, broker, connected)
		case <-sig:
			showCursor()
			fmt.Println("\nAu revoir.")
			return
		}
	}
}
