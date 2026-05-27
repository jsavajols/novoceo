package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

var paris = func() *time.Location {
	loc, _ := time.LoadLocation("Europe/Paris")
	return loc
}()

func logf(format string, args ...any) {
	fmt.Printf("[front] "+format+"\n", args...)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ---- App ----

type app struct {
	apiURL string
	token  string
	client *http.Client
}

func (a *app) get(path string, out interface{}) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, a.apiURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API %d: %s", resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (a *app) post(path, body string) error {
	req, err := http.NewRequestWithContext(
		context.Background(), http.MethodPost, a.apiURL+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	return nil
}

// ---- Helpers partagés ----

func fmtTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(paris).Format("02/01 15:04")
}

func pf(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func pi(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func ps(s *string) string {
	if s == nil {
		return "?"
	}
	return *s
}

// ---- Main ----

func main() {
	_ = godotenv.Load()

	port := env("FRONT_PORT", "3000")
	apiURL := env("API_URL", "http://api:5000")
	token := env("API_TOKEN", "")
	if token == "" {
		fmt.Fprintln(os.Stderr, "[front] API_TOKEN non défini — arrêt")
		os.Exit(1)
	}

	a := &app{
		apiURL: apiURL,
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}

	f := fiber.New(fiber.Config{AppName: "novoceo-front"})

	f.Get("/", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(indexHTML)
	})
	f.Get("/htmx/watchdog", a.htmxWatchdog)
	f.Get("/htmx/health", a.htmxHealth)
	f.Get("/htmx/temperature", a.htmxTemperature)
	f.Get("/htmx/prise", a.htmxPrise)
	f.Post("/htmx/toggle", a.htmxToggle)
	f.Get("/htmx/contact", a.htmxContact)
	f.Get("/htmx/presence", a.htmxPresence)

	logf("démarré sur :%s — API: %s", port, apiURL)
	if err := f.Listen(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "[front] %v\n", err)
		os.Exit(1)
	}
}
