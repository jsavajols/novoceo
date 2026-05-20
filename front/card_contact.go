package main

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type contactData struct {
	Contact   *bool     `json:"contact"`
	Battery   *int      `json:"battery"`
	CreatedAt time.Time `json:"created_at"`
}

func (a *app) htmxContact(c *fiber.Ctx) error {
	var d contactData
	_ = a.get("/device/Contacteur%20porte/contact", &d)

	var label, glowClass, textColor, dotColor string
	switch {
	case d.Contact == nil:
		label, glowClass, textColor, dotColor = "?", "", "text-slate-500", "bg-slate-600"
	case *d.Contact:
		label, glowClass, textColor, dotColor = "FERMÉ", "glow-off", "text-violet-300", "bg-violet-400"
	default:
		label, glowClass, textColor, dotColor = "OUVERT", "glow-alert", "text-red-400", "bg-red-500"
	}

	bat := ""
	if d.Battery != nil {
		bat = fmt.Sprintf(
			`<div><p class="font-mono text-base text-slate-500 mb-0.5">batterie</p>`+
				`<p class="font-mono text-base text-green-300">%d%%</p></div>`,
			*d.Battery,
		)
	}

	html := fmt.Sprintf(`<div class="space-y-4">
  <div class="flex flex-col items-center gap-3 py-2">
    <div class="relative">
      <div class="h-24 w-24 rounded-full bg-slate-800 border-2 border-slate-700 %s flex items-center justify-center transition-all duration-700">
        <span class="font-mono text-base font-medium tracking-tight %s">%s</span>
      </div>
      <span class="absolute bottom-0.5 right-0.5 h-3.5 w-3.5 rounded-full border-2 border-slate-900 %s"></span>
    </div>
    %s
    <p class="font-mono text-xs text-slate-500">%s</p>
  </div>
</div>`,
		glowClass, textColor, label, dotColor,
		bat,
		fmtTime(d.CreatedAt),
	)

	return c.Type("html").SendString(html)
}
