package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type watchdogData struct {
	Loss      int       `json:"loss"`
	CreatedAt time.Time `json:"created_at"`
	OK        bool      `json:"ok"`
}

const dotAttrs = `id="watchdog-dot" hx-get="/htmx/watchdog" hx-trigger="every 60s" hx-swap="outerHTML"`

func (a *app) htmxWatchdog(c *fiber.Ctx) error {
	var data watchdogData
	err := a.get("/watchdog", &data)

	var color, title string
	if err != nil {
		color = "bg-red-500"
		title = "RPi injoignable"
	} else if !data.OK {
		color = "bg-red-500"
		title = fmt.Sprintf("RPi ALERTE - perte %d%%", data.Loss)
	} else {
		color = "bg-emerald-400"
		title = fmt.Sprintf("RPi OK - perte %d%%", data.Loss)
	}

	return c.Type("html").SendString(fmt.Sprintf(
		`<span %s class="ml-1 h-3 w-3 rounded-full %s animate-pulse" title="%s"></span>`,
		dotAttrs, color, title,
	))
}
