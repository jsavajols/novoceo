// api : service REST exposant les données Zigbee2MQTT et le contrôle des devices.
//
//	POST /device/command        — envoie une commande MQTT (équivalent de send-device)
//	GET  /sensor/temperature    — dernière mesure du capteur Température
//	GET  /bridge/health         — dernier état de santé du bridge Zigbee2MQTT
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func log(format string, args ...any) {
	fmt.Printf("[api] "+format+"\n", args...)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type application struct {
	db   *pgxpool.Pool
	mqtt mqtt.Client
}

// --- POST /device/command ---

type commandRequest struct {
	Device  string `json:"device"`
	RPI     string `json:"rpi"`
	State   string `json:"state"`
	Payload string `json:"payload"`
}

func (a *application) postCommand(c *fiber.Ctx) error {
	var req commandRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if req.Device == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "device est requis"})
	}

	var topic string
	if req.RPI != "" {
		topic = fmt.Sprintf("zigbee2mqtt/%s/%s/set", req.RPI, req.Device)
	} else {
		topic = fmt.Sprintf("zigbee2mqtt/%s/set", req.Device)
	}

	var msg string
	if req.Payload != "" {
		var check map[string]interface{}
		if err := json.Unmarshal([]byte(req.Payload), &check); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "payload JSON invalide"})
		}
		msg = req.Payload
	} else {
		state := req.State
		if state == "" {
			state = "TOGGLE"
		}
		msg = fmt.Sprintf(`{"state":"%s"}`, state)
	}

	if !a.mqtt.IsConnected() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "broker MQTT non disponible"})
	}
	token := a.mqtt.Publish(topic, 0, false, msg)
	if !token.WaitTimeout(3 * time.Second) {
		return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{"error": "timeout broker MQTT"})
	}
	if err := token.Error(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"ok": true, "topic": topic, "payload": msg})
}

// --- GET /sensor/temperature ---

type temperatureResponse struct {
	Battery     *int      `json:"battery"`
	Temperature *float64  `json:"temperature"`
	Humidity    *float64  `json:"humidity"`
	CreatedAt   time.Time `json:"created_at"`
}

const queryTemperature = `
SELECT
    (device_state->>'battery')::int       AS battery,
    (device_state->>'temperature')::float AS temperature,
    (device_state->>'humidity')::float    AS humidity,
    created_at
FROM states
WHERE topic LIKE '%Température%'
ORDER BY id DESC
LIMIT 1`

func (a *application) getTemperature(c *fiber.Ctx) error {
	var res temperatureResponse
	row := a.db.QueryRow(context.Background(), queryTemperature)
	err := row.Scan(&res.Battery, &res.Temperature, &res.Humidity, &res.CreatedAt)
	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aucune donnée"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// --- GET /sensor/temperature/history ---

type historyPoint struct {
	Temperature float64   `json:"temperature"`
	CreatedAt   time.Time `json:"created_at"`
}

const queryTemperatureHistory = `
SELECT
    AVG((device_state->>'temperature')::float)::float AS temperature,
    date_trunc('hour', created_at) +
        (EXTRACT(minute FROM created_at)::int / 15 * 15) * interval '1 minute' AS bucket
FROM states
WHERE topic LIKE '%Température%'
  AND created_at > NOW() - INTERVAL '24 hours'
  AND (device_state->>'temperature') IS NOT NULL
GROUP BY bucket
ORDER BY bucket ASC`

func (a *application) getTemperatureHistory(c *fiber.Ctx) error {
	rows, err := a.db.Query(context.Background(), queryTemperatureHistory)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	points := []historyPoint{}
	for rows.Next() {
		var p historyPoint
		if err := rows.Scan(&p.Temperature, &p.CreatedAt); err != nil {
			continue
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(points)
}

// --- GET /bridge/health ---

type healthResponse struct {
	LoadAverage   json.RawMessage `json:"load_average"`
	MemoryPercent *float64        `json:"memory_percent"`
	MemoryFree    *float64        `json:"memory_free"`
	CreatedAt     time.Time       `json:"created_at"`
}

const queryHealth = `
SELECT
    device_state->'os'->'load_average'                        AS load_average,
    (device_state->'os'->>'memory_percent')::float            AS memory_percent,
    (100 - (device_state->'os'->>'memory_percent')::float)    AS memory_free,
    created_at
FROM states
WHERE topic = 'zigbee2mqtt/bridge/health'
ORDER BY id DESC
LIMIT 1`

func (a *application) getBridgeHealth(c *fiber.Ctx) error {
	var res healthResponse
	var loadAvgRaw []byte
	row := a.db.QueryRow(context.Background(), queryHealth)
	err := row.Scan(&loadAvgRaw, &res.MemoryPercent, &res.MemoryFree, &res.CreatedAt)
	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aucune donnée"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	res.LoadAverage = json.RawMessage(loadAvgRaw)
	return c.JSON(res)
}

// --- GET /device/:device/contact ---

type contactResponse struct {
	Contact   *bool     `json:"contact"`
	Battery   *int      `json:"battery"`
	CreatedAt time.Time `json:"created_at"`
}

const queryContactState = `
SELECT
    (device_state->>'contact')::boolean AS contact,
    (device_state->>'battery')::int     AS battery,
    created_at
FROM states
WHERE topic LIKE '%' || $1 || '%'
  AND topic NOT LIKE '%/set%'
  AND device_state->>'contact' IS NOT NULL
ORDER BY id DESC
LIMIT 1`

func (a *application) getContactState(c *fiber.Ctx) error {
	device, _ := url.PathUnescape(c.Params("device"))
	var res contactResponse
	row := a.db.QueryRow(context.Background(), queryContactState, device)
	err := row.Scan(&res.Contact, &res.Battery, &res.CreatedAt)
	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aucune donnée"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// --- GET /device/:device/state ---

const queryDeviceState = `
SELECT
    device_state->>'state' AS state,
    created_at
FROM states
WHERE topic LIKE '%' || $1 || '%'
  AND topic NOT LIKE '%/set%'
ORDER BY id DESC
LIMIT 1`

type deviceStateResponse struct {
	State     *string   `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *application) getDeviceState(c *fiber.Ctx) error {
	device, _ := url.PathUnescape(c.Params("device"))
	var res deviceStateResponse
	row := a.db.QueryRow(context.Background(), queryDeviceState, device)
	err := row.Scan(&res.State, &res.CreatedAt)
	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "aucune donnée"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(res)
}

// --- Init ---

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", ""),
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_NAME", "postgres"),
	)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	log("connecté à PostgreSQL (%s:%s/%s)", env("DB_HOST", "localhost"), env("DB_PORT", "5432"), env("DB_NAME", "postgres"))
	return pool, nil
}

func initMQTT(broker string) mqtt.Client {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(fmt.Sprintf("api-%d", time.Now().Unix())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(3 * time.Second).
		SetOnConnectHandler(func(_ mqtt.Client) {
			log("MQTT connecté à %s", broker)
		}).
		SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			log("MQTT connexion perdue : %v — reconnexion...", err)
		})

	client := mqtt.NewClient(opts)
	// Connexion non-bloquante : l'API démarre immédiatement, MQTT se connecte en arrière-plan.
	// postCommand vérifie IsConnected() avant de publier.
	client.Connect()
	return client
}

func bearerAuth(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Get("Authorization") != "Bearer "+token {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
		}
		return c.Next()
	}
}

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	broker := env("MQTT_BROKER", "tcp://192.168.1.128:1883")
	port := env("API_PORT", "5000")
	token := env("API_TOKEN", "")
	if token == "" {
		fmt.Fprintln(os.Stderr, "[api] API_TOKEN non défini — arrêt")
		os.Exit(1)
	}

	db, err := initDB(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[api] PostgreSQL connexion échouée : %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	mqttClient := initMQTT(broker)
	defer mqttClient.Disconnect(250)

	a := &application{db: db, mqtt: mqttClient}

	f := fiber.New(fiber.Config{
		AppName: "novoceo-api",
	})

	f.Get("/healthz", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	f.Use(bearerAuth(token))

	f.Post("/device/command", a.postCommand)
	f.Get("/device/:device/contact", a.getContactState)
	f.Get("/device/:device/state", a.getDeviceState)
	f.Get("/sensor/temperature", a.getTemperature)
	f.Get("/sensor/temperature/history", a.getTemperatureHistory)
	f.Get("/bridge/health", a.getBridgeHealth)

	log("démarrée sur :%s", port)
	if err := f.Listen(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "[api] %v\n", err)
		os.Exit(1)
	}
}
