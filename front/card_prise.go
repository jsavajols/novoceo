package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type priseData struct {
	State     *string   `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *app) renderPrise(c *fiber.Ctx) error {
	var d priseData
	_ = a.get("/device/Prise/state", &d)

	state := ps(d.State)
	on := strings.EqualFold(state, "ON")

	glowClass := "glow-off"
	textColor := "text-violet-400"
	dotColor := "bg-violet-500"
	if on {
		glowClass = "glow-on"
		textColor = "text-emerald-400"
		dotColor = "bg-emerald-400"
	}

	html := fmt.Sprintf(`<div class="space-y-5">
  <div class="flex flex-col items-center gap-3 py-2">
    <div class="relative">
      <div class="h-24 w-24 rounded-full bg-slate-800 border-2 border-slate-700 %s flex items-center justify-center transition-all duration-700">
        <span class="font-mono text-xl font-medium %s">%s</span>
      </div>
      <span class="absolute bottom-0.5 right-0.5 h-3.5 w-3.5 rounded-full border-2 border-slate-900 %s"></span>
    </div>
    <p class="font-mono text-xs text-slate-700">%s</p>
  </div>
  <button
    hx-post="/htmx/toggle"
    hx-target="#prise-content"
    hx-swap="innerHTML"
    class="w-full py-2.5 rounded-xl font-mono text-sm font-medium bg-slate-800 hover:bg-slate-700 active:scale-95 border border-slate-700 hover:border-slate-600 text-slate-300 transition-all duration-150 cursor-pointer">
    Toggle
  </button>
</div>`,
		glowClass, textColor, state, dotColor,
		fmtTime(d.CreatedAt),
	)

	return c.Type("html").SendString(html)
}

func (a *app) htmxPrise(c *fiber.Ctx) error {
	return a.renderPrise(c)
}

func (a *app) htmxToggle(c *fiber.Ctx) error {
	_ = a.post("/device/command", `{"device":"Prise","state":"TOGGLE"}`)
	time.Sleep(1500 * time.Millisecond)
	return a.renderPrise(c)
}
