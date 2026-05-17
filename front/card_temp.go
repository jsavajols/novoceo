package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type tempData struct {
	Battery     *int      `json:"battery"`
	Temperature *float64  `json:"temperature"`
	Humidity    *float64  `json:"humidity"`
	CreatedAt   time.Time `json:"created_at"`
}

type historyPoint struct {
	Temperature float64   `json:"temperature"`
	CreatedAt   time.Time `json:"created_at"`
}

func sparklineSVG(points []historyPoint) (svg string, minT, maxT float64) {
	if len(points) < 2 {
		return "", 0, 0
	}

	const svgW, svgH = 260.0, 50.0
	const pad = 3.0

	minT = points[0].Temperature
	maxT = points[0].Temperature
	for _, p := range points {
		if p.Temperature < minT {
			minT = p.Temperature
		}
		if p.Temperature > maxT {
			maxT = p.Temperature
		}
	}

	rangeT := maxT - minT
	if rangeT < 0.1 {
		rangeT = 1.0
	}

	n := len(points)
	coords := make([]string, n)
	for i, p := range points {
		x := float64(i) / float64(n-1) * svgW
		y := svgH - pad - (p.Temperature-minT)/rangeT*(svgH-2*pad)
		coords[i] = fmt.Sprintf("%.1f,%.1f", x, y)
	}

	line := strings.Join(coords, " ")
	area := fmt.Sprintf("0,%.1f %s %.1f,%.1f", svgH, line, svgW, svgH)

	svg = `<svg viewBox="0 0 260 50" preserveAspectRatio="none" class="w-full" style="height:48px">` +
		`<defs><linearGradient id="tg" x1="0" y1="0" x2="0" y2="1">` +
		`<stop offset="0%" stop-color="#f59e0b" stop-opacity="0.3"/>` +
		`<stop offset="100%" stop-color="#f59e0b" stop-opacity="0"/>` +
		`</linearGradient></defs>` +
		`<polygon points="` + area + `" fill="url(#tg)"/>` +
		`<polyline points="` + line + `" fill="none" stroke="#d97706" stroke-width="1.5" stroke-linejoin="round" stroke-linecap="round"/>` +
		`</svg>`

	return svg, minT, maxT
}

func (a *app) htmxTemperature(c *fiber.Ctx) error {
	var d tempData
	if err := a.get("/sensor/temperature", &d); err != nil {
		return c.Type("html").SendString(`<p class="font-mono text-xs text-red-400">capteur indisponible</p>`)
	}

	var history []historyPoint
	_ = a.get("/sensor/temperature/history", &history)

	var chart string
	if svg, minT, maxT := sparklineSVG(history); svg != "" {
		chart = `<div class="border-t border-slate-800 pt-3 space-y-1.5">` +
			`<div class="flex justify-between items-center">` +
			`<span class="font-mono text-xs text-slate-500">24 dernières heures</span>` +
			fmt.Sprintf(`<span class="font-mono text-xs text-amber-800">↓%.1f° ↑%.1f°</span>`, minT, maxT) +
			`</div>` +
			svg +
			`<div class="flex justify-between">` +
			`<span class="font-mono text-xs text-slate-700">-24h</span>` +
			`<span class="font-mono text-xs text-slate-700">now</span>` +
			`</div></div>`
	} else {
		chart = `<div class="border-t border-slate-800 pt-3">` +
			`<p class="font-mono text-xs text-slate-500 mb-1">24 dernières heures</p>` +
			`<p class="font-mono text-xs text-slate-700">aucune donnée</p></div>`
	}

	html := fmt.Sprintf(`<div class="space-y-4">
  <div class="flex items-end gap-1.5">
    <span class="font-mono text-5xl font-medium text-amber-300 leading-none">%.1f</span>
    <span class="font-mono text-lg text-amber-500 mb-1">°C</span>
  </div>
  <div class="flex gap-5">
    <div>
      <p class="font-mono text-xs text-slate-500 mb-0.5">humidité</p>
      <p class="font-mono text-sm text-blue-300">%.0f%%</p>
    </div>
    <div>
      <p class="font-mono text-xs text-slate-500 mb-0.5">batterie</p>
      <p class="font-mono text-sm text-green-300">%d%%</p>
    </div>
  </div>
  %s
  <p class="font-mono text-xs text-slate-700">%s</p>
</div>`,
		pf(d.Temperature),
		pf(d.Humidity),
		pi(d.Battery),
		chart,
		fmtTime(d.CreatedAt),
	)

	return c.Type("html").SendString(html)
}
