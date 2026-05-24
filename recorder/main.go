// recorder : souscrit à tous les topics zigbee2mqtt/# et persiste chaque payload
// dans la table PostgreSQL `states`. Configuration 100% via variables d'environnement
// pour un déploiement K8s (Secret/ConfigMap) ; le fichier .env sert uniquement en local.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func log(format string, args ...any) {
	fmt.Printf("[recorder] "+format+"\n", args...)
}

func logErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[recorder] ERROR "+format+"\n", args...)
}

type config struct {
	mqttBroker string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPassword string
	dbName     string
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loadConfig() config {
	// .env ignoré silencieusement en production : les vars sont injectées par K8s.
	_ = godotenv.Load()
	return config{
		// ancien défaut local : env("MQTT_BROKER", "tcp://localhost:1883")
		mqttBroker: env("MQTT_BROKER", "tcp://mosquitto:1883"),
		dbHost:     env("DB_HOST", "localhost"),
		dbPort:     env("DB_PORT", "5432"),
		dbUser:     env("DB_USER", "postgres"),
		dbPassword: env("DB_PASSWORD", ""),
		dbName:     env("DB_NAME", "postgres"),
	}
}

const insertState = `INSERT INTO states (topic, device_state) VALUES ($1, $2)`

func initDB(ctx context.Context, cfg config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.dbUser, cfg.dbPassword, cfg.dbHost, cfg.dbPort, cfg.dbName,
	)

	var (
		pool *pgxpool.Pool
		err  error
	)

	// Backoff quadratique (1s, 4s, 9s, 16s, 25s) puis exit : K8s redémarre le pod.
	for attempt := 1; attempt <= 5; attempt++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				break
			} else {
				pool.Close()
				err = pingErr
			}
		}
		wait := time.Duration(attempt*attempt) * time.Second
		logErr("connexion DB échouée (tentative %d/5) : %v — retry dans %s", attempt, err, wait)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("impossible de se connecter à PostgreSQL après 5 tentatives : %w", err)
	}

	log("connecté à PostgreSQL (%s:%s/%s)", cfg.dbHost, cfg.dbPort, cfg.dbName)
	return pool, nil
}

// sanitizePayload supprime \u0000 et null bytes que PostgreSQL JSONB rejette (SQLSTATE 22P05).
func sanitizePayload(raw []byte) []byte {
	raw = bytes.ReplaceAll(raw, []byte("\\u0000"), []byte{})
	raw = bytes.ReplaceAll(raw, []byte{0x00}, []byte{})
	return raw
}

// normalizePayload wrap les payloads non-JSON (ex: "ON", "1024") en {"raw":"..."}
// pour que la colonne JSONB reste uniforme et requêtable.
func normalizePayload(raw []byte) json.RawMessage {
	raw = sanitizePayload(raw)
	if json.Valid(raw) {
		return json.RawMessage(raw)
	}
	wrapped, _ := json.Marshal(map[string]string{"raw": string(raw)})
	return json.RawMessage(wrapped)
}

func makeHandler(pool *pgxpool.Pool) mqtt.MessageHandler {
	return func(_ mqtt.Client, msg mqtt.Message) {
		topic := msg.Topic()
		payload := normalizePayload(msg.Payload())

		// Timeout par message : un pic de latence DB ne bloque pas la réception MQTT.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := pool.Exec(ctx, insertState, topic, payload); err != nil {
			logErr("insert échoué [%s] : %v", topic, err)
			return
		}
		log("stocké : %s", topic)
	}
}

func connectMQTT(cfg config, pool *pgxpool.Pool) mqtt.Client {
	handler := makeHandler(pool)

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.mqttBroker).
		SetClientID(fmt.Sprintf("recorder-%d", time.Now().Unix())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log("connecté à %s", cfg.mqttBroker)
			// Subscribe dans OnConnectHandler : la souscription est rejouée après chaque reconnexion.
			if token := c.Subscribe("zigbee2mqtt/#", 1, handler); token.Wait() && token.Error() != nil {
				logErr("subscribe échoué : %v", token.Error())
			} else {
				log("abonné à zigbee2mqtt/#")
			}
			if token := c.Subscribe("rpi/#", 1, handler); token.Wait() && token.Error() != nil {
				logErr("subscribe échoué : %v", token.Error())
			} else {
				log("abonné à rpi/#")
			}
			if token := c.Subscribe("rpi0/#", 1, handler); token.Wait() && token.Error() != nil {
				logErr("subscribe échoué : %v", token.Error())
			} else {
				log("abonné à rpi0/#")
			}
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			logErr("connexion MQTT perdue : %v — reconnexion automatique...", err)
		})

	return mqtt.NewClient(opts)
}

func main() {
	cfg := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := initDB(ctx, cfg)
	if err != nil {
		logErr("%v", err)
		os.Exit(1)
	}
	defer pool.Close()

	client := connectMQTT(cfg, pool)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logErr("connexion MQTT initiale échouée : %v", token.Error())
		os.Exit(1)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	received := <-sig

	log("signal reçu (%s) — arrêt propre...", received)
	cancel()
	client.Disconnect(500)
	log("arrêt.")
}
