package main

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v2"
)

type healthData struct {
	LoadAverage   json.RawMessage `json:"load_average"`
	MemoryPercent *float64        `json:"memory_percent"`
	CreatedAt     time.Time       `json:"created_at"`
}

func (a *app) htmxHealth(c *fiber.Ctx) error {
	var d healthData
	if err := a.get("/bridge/health", &d); err != nil {
		return c.Type("html").SendString(`<p class="font-mono text-xs text-red-400">bridge indisponible</p>`)
	}

	var loads [3]float64
	_ = json.Unmarshal(d.LoadAverage, &loads)

	scale := math.Max(loads[0], math.Max(loads[1], loads[2]))
	if scale < 0.01 {
		scale = 1.0
	}
	pct := func(v float64) float64 {
		p := v / scale * 90
		if p < 4 {
			p = 4
		}
		return p
	}

	mem := pf(d.MemoryPercent)

	html := fmt.Sprintf(`<div class="space-y-4">
  <p class="font-mono text-base text-slate-500 mb-1">load average</p>
  <div class="flex gap-1.5">
    <span class="flex-1 text-center font-mono text-sm text-slate-300">%.2f</span>
    <span class="flex-1 text-center font-mono text-sm text-slate-300">%.2f</span>
    <span class="flex-1 text-center font-mono text-sm text-slate-300">%.2f</span>
  </div>
  <div class="flex items-end gap-1.5" style="height:48px">
    <div class="flex-1 bg-cyan-500 rounded-t-sm" style="height:%.0f%%"></div>
    <div class="flex-1 bg-cyan-600 rounded-t-sm" style="height:%.0f%%"></div>
    <div class="flex-1 bg-cyan-700 rounded-t-sm" style="height:%.0f%%"></div>
  </div>
  <div class="flex gap-1.5 mb-1">
    <span class="flex-1 text-center font-mono text-base text-slate-500">1m</span>
    <span class="flex-1 text-center font-mono text-base text-slate-500">5m</span>
    <span class="flex-1 text-center font-mono text-base text-slate-500">15m</span>
  </div>
  <div class="space-y-2 pt-2 border-t border-slate-800">
    <div class="flex justify-between">
      <span class="font-mono text-base text-slate-500">mémoire</span>
      <span class="font-mono text-base text-cyan-400">%.0f%%</span>
    </div>
    <div class="h-2 bg-slate-800 rounded-full overflow-hidden">
      <div class="h-full bg-gradient-to-r from-cyan-600 to-cyan-400 rounded-full" style="width:%.0f%%"></div>
    </div>
  </div>
  <p class="font-mono text-xs text-slate-500">%s</p>
</div>`,
		loads[0], loads[1], loads[2],
		pct(loads[0]), pct(loads[1]), pct(loads[2]),
		mem, mem,
		fmtTime(d.CreatedAt),
	)

	return c.Type("html").SendString(html)
}
