package main

const indexHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>NOVOCEO · IoT</title>
  <link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%23020617'/><text x='16' y='23' text-anchor='middle' font-family='monospace' font-size='20' font-weight='700' fill='%232563EB'>n</text></svg>">
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link href="https://fonts.googleapis.com/css2?family=DM+Mono:wght@300;400;500&family=DM+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet">
  <script src="https://cdn.tailwindcss.com"></script>
  <script>
    tailwind.config = {
      theme: {
        extend: {
          fontFamily: {
            mono: ['DM Mono', 'monospace'],
            sans: ['DM Sans', 'sans-serif'],
          }
        }
      }
    }
  </script>
  <script src="https://unpkg.com/htmx.org@2.0.3/dist/htmx.min.js"></script>
  <style>
    /* ── Dark (défaut) ── */
    body { background: #020617; }
    .card-cyan   { box-shadow: 0 0 0 1px rgba(6,182,212,.1),   0 8px 32px rgba(6,182,212,.06); }
    .card-amber  { box-shadow: 0 0 0 1px rgba(245,158,11,.1),  0 8px 32px rgba(245,158,11,.06); }
    .card-emer   { box-shadow: 0 0 0 1px rgba(52,211,153,.1),  0 8px 32px rgba(52,211,153,.06); }
    .card-violet { box-shadow: 0 0 0 1px rgba(139,92,246,.1),  0 8px 32px rgba(139,92,246,.06); }
    .card-sky    { box-shadow: 0 0 0 1px rgba(56,189,248,.1),   0 8px 32px rgba(56,189,248,.06); }
    .glow-on    { box-shadow: 0 0 30px rgba(52,211,153,.4),  0 0 80px rgba(52,211,153,.12);  border-color: rgba(52,211,153,.35)  !important; }
    .glow-off   { box-shadow: 0 0 20px rgba(139,92,246,.25), 0 0 60px rgba(139,92,246,.08);  border-color: rgba(139,92,246,.25)  !important; }
    .glow-alert { box-shadow: 0 0 30px rgba(239,68,68,.4),   0 0 80px rgba(239,68,68,.12);   border-color: rgba(239,68,68,.35)   !important; }
    .htmx-indicator { opacity: 0; transition: opacity .2s ease-in; }
    .htmx-request .htmx-indicator { opacity: 1; }
    .htmx-request.htmx-indicator  { opacity: 1; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .spin { animation: spin .8s linear infinite; display: inline-block; }

    /* ── Sélecteur de thème ── */
    .theme-pill {
      width: 28px; height: 28px; border-radius: 50%;
      display: flex; align-items: center; justify-content: center;
      font-size: 13px; line-height: 1;
      background: #1e293b; border: 1px solid #334155; color: #475569;
      cursor: pointer; transition: background .15s, border-color .15s, color .15s;
    }
    .theme-pill.active { background: #2563eb; border-color: #2563eb; color: #f1f5f9; }

    /* ── Light ── */
    html.light body              { background: #f1f5f9; }
    html.light .bg-slate-900     { background-color: #ffffff !important; }
    html.light .bg-slate-800     { background-color: #e2e8f0 !important; }
    html.light .text-slate-100   { color: #0f172a   !important; }
    html.light .text-slate-300   { color: #1e293b   !important; }
    html.light .text-slate-400   { color: #334155   !important; }
    html.light .text-slate-500   { color: #475569   !important; }
    html.light .text-slate-600   { color: #64748b   !important; }
    html.light .text-slate-700   { color: #94a3b8   !important; }
    html.light .border-slate-700 { border-color: #cbd5e1 !important; }
    html.light .border-slate-800 { border-color: #e2e8f0 !important; }
    html.light .border-slate-900 { border-color: #e2e8f0 !important; }
    /* Couleurs d'accent : assombrir pour contraste sur fond clair */
    html.light .text-amber-300  { color: #d97706 !important; }
    html.light .text-amber-400  { color: #b45309 !important; }
    html.light .text-amber-500  { color: #92400e !important; }
    html.light .text-blue-300   { color: #1d4ed8 !important; }
    html.light .text-green-300  { color: #15803d !important; }
    html.light .text-green-400  { color: #16a34a !important; }
    html.light .text-cyan-400   { color: #0e7490 !important; }
    html.light .text-emerald-400 { color: #047857 !important; }
    html.light .text-violet-300 { color: #6d28d9 !important; }
    html.light .text-violet-400 { color: #6d28d9 !important; }
    html.light .text-red-400    { color: #b91c1c !important; }
    /* Tintes de fond par carte : identité visuelle en mode clair */
    html.light .card-cyan   { background-color: #ecfeff !important; box-shadow: 0 0 0 1px rgba(6,182,212,.3),  0 4px 24px rgba(6,182,212,.1); }
    html.light .card-amber  { background-color: #fffbeb !important; box-shadow: 0 0 0 1px rgba(245,158,11,.3), 0 4px 24px rgba(245,158,11,.1); }
    html.light .card-emer   { background-color: #f0fdf4 !important; box-shadow: 0 0 0 1px rgba(52,211,153,.3), 0 4px 24px rgba(52,211,153,.1); }
    html.light .card-violet { background-color: #f5f3ff !important; box-shadow: 0 0 0 1px rgba(139,92,246,.3), 0 4px 24px rgba(139,92,246,.1); }
    html.light .card-sky    { background-color: #f0f9ff !important; box-shadow: 0 0 0 1px rgba(56,189,248,.3),  0 4px 24px rgba(56,189,248,.1); }
    html.light .text-sky-300 { color: #0284c7 !important; }
    html.light .text-sky-400 { color: #0369a1 !important; }
    html.light .glow-on    { box-shadow: 0 0 20px rgba(52,211,153,.3),  0 0 50px rgba(52,211,153,.08);  border-color: rgba(52,211,153,.5)  !important; }
    html.light .glow-off   { box-shadow: 0 0 15px rgba(139,92,246,.2),  0 0 40px rgba(139,92,246,.06); border-color: rgba(139,92,246,.5)  !important; }
    html.light .glow-alert { box-shadow: 0 0 20px rgba(239,68,68,.3),   0 0 50px rgba(239,68,68,.08);  border-color: rgba(239,68,68,.5)   !important; }
    html.light .theme-pill { background: #e2e8f0; border-color: #cbd5e1; color: #64748b; }
    html.light .hover\:bg-slate-700:hover  { background-color: #e2e8f0 !important; }
    html.light .hover\:border-slate-600:hover { border-color: #cbd5e1 !important; }
  </style>
  <script>
    (function(){
      var mode = localStorage.getItem('theme') || 'system';
      var resolved = mode === 'system'
        ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
        : mode;
      if (resolved === 'light') document.documentElement.classList.add('light');
    })();
  </script>
</head>
<body class="font-sans text-slate-100 min-h-screen p-4 md:p-8">

  <header class="mb-8 flex items-center justify-between max-w-6xl mx-auto">
    <!-- Logo NOVOCEO GROUPE -->
    <div>
      <div class="flex items-baseline" style="letter-spacing:-0.04em">
        <span class="font-sans font-bold text-xl">NOVO</span><span class="font-sans font-bold text-xl" style="color:#2563EB">CEO</span>
      </div>
      <div class="flex items-center gap-1.5" style="margin-top:3px">
        <div style="height:1px;width:16px;background:#2563EB"></div>
        <span class="font-mono text-slate-600" style="font-size:8px;letter-spacing:0.32em">GROUPE</span>
        <div style="height:1px;width:16px;background:#2563EB"></div>
      </div>
    </div>
    <!-- Contrôles -->
    <div class="flex items-center gap-2">
      <button class="theme-pill" id="btn-light"  onclick="setTheme('light')"  title="Mode clair">&#9728;</button>
      <button class="theme-pill" id="btn-system" onclick="setTheme('system')" title="Mode système">&#9680;</button>
      <button class="theme-pill" id="btn-dark"   onclick="setTheme('dark')"   title="Mode sombre">&#9790;</button>
      <span id="watchdog-dot" hx-get="/htmx/watchdog" hx-trigger="load, every 60s" hx-swap="outerHTML" class="ml-1 h-3 w-3 rounded-full bg-slate-700 animate-pulse" title="RPi..."></span>
    </div>
  </header>

  <main class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-5 max-w-6xl mx-auto">

    <!-- Bridge Health — refresh 60s -->
    <div class="bg-slate-900 rounded-2xl p-5 card-cyan">
      <div class="flex items-center gap-2 mb-5">
        <svg class="w-3.5 h-3.5 text-cyan-400 shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zm6-4a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zm6-3a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z"/>
        </svg>
        <span class="font-mono text-xs text-cyan-400 tracking-widest uppercase">Bridge Health</span>
        <span id="health-spinner" class="htmx-indicator ml-auto">
          <span class="spin inline-block h-3 w-3 rounded-full border-2 border-slate-700 border-t-cyan-400"></span>
        </span>
      </div>
      <div id="health-content"
        hx-get="/htmx/health"
        hx-trigger="load, every 60s"
        hx-swap="innerHTML"
        hx-indicator="#health-spinner">
        <p class="font-mono text-xs text-slate-500">chargement...</p>
      </div>
    </div>

    <!-- Temperature — refresh 60s -->
    <div class="bg-slate-900 rounded-2xl p-5 card-amber">
      <div class="flex items-center gap-2 mb-5">
        <svg class="w-3.5 h-3.5 text-amber-400 shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v4a4 4 0 108 0V6a4 4 0 00-4-4zm0 14a6 6 0 100-12 6 6 0 000 12z" clip-rule="evenodd"/>
        </svg>
        <span class="font-mono text-xs text-amber-400 tracking-widest uppercase">Température</span>
        <span id="temp-spinner" class="htmx-indicator ml-auto">
          <span class="spin inline-block h-3 w-3 rounded-full border-2 border-slate-700 border-t-amber-400"></span>
        </span>
      </div>
      <div id="temperature-content"
        hx-get="/htmx/temperature"
        hx-trigger="load, every 60s"
        hx-swap="innerHTML"
        hx-indicator="#temp-spinner">
        <p class="font-mono text-xs text-slate-500">chargement...</p>
      </div>
    </div>

    <!-- Prise — refresh 5s -->
    <div class="bg-slate-900 rounded-2xl p-5 card-emer">
      <div class="flex items-center gap-2 mb-5">
        <svg class="w-3.5 h-3.5 text-emerald-400 shrink-0" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M11.3 1.046A1 1 0 0112 2v5h4a1 1 0 01.82 1.573l-7 10A1 1 0 018 18v-5H4a1 1 0 01-.82-1.573l7-10a1 1 0 011.12-.38z" clip-rule="evenodd"/>
        </svg>
        <span class="font-mono text-xs text-emerald-400 tracking-widest uppercase">Prise</span>
      </div>
      <div id="prise-content"
        hx-get="/htmx/prise"
        hx-trigger="load, every 5s"
        hx-swap="innerHTML">
        <p class="font-mono text-xs text-slate-500">chargement...</p>
      </div>
    </div>

    <!-- Porte — refresh 5s -->
    <div class="bg-slate-900 rounded-2xl p-5 card-violet">
      <div class="flex items-center gap-2 mb-5">
        <svg class="w-3.5 h-3.5 text-violet-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
        </svg>
        <span class="font-mono text-xs text-violet-400 tracking-widest uppercase">Porte</span>
        <span id="contact-spinner" class="htmx-indicator ml-auto">
          <span class="spin inline-block h-3 w-3 rounded-full border-2 border-slate-700 border-t-violet-400"></span>
        </span>
      </div>
      <div id="contact-content"
        hx-get="/htmx/contact"
        hx-trigger="load, every 5s"
        hx-swap="innerHTML"
        hx-indicator="#contact-spinner">
        <p class="font-mono text-xs text-slate-500">chargement...</p>
      </div>
    </div>


    <!-- Présence rpi0 — refresh 5s -->
    <div class="bg-slate-900 rounded-2xl p-5 card-sky">
      <div class="flex items-center gap-2 mb-5">
        <svg class="w-3.5 h-3.5 text-sky-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z"/>
        </svg>
        <span class="font-mono text-xs text-sky-400 tracking-widest uppercase">Présence</span>
        <span id="presence-spinner" class="htmx-indicator ml-auto">
          <span class="spin inline-block h-3 w-3 rounded-full border-2 border-slate-700 border-t-sky-400"></span>
        </span>
      </div>
      <div id="presence-content"
        hx-get="/htmx/presence"
        hx-trigger="load, every 2s"
        hx-swap="innerHTML"
        hx-indicator="#presence-spinner">
        <p class="font-mono text-xs text-slate-500">chargement...</p>
      </div>
    </div>

  </main>

  <footer class="mt-10 text-center font-mono text-xs text-slate-500 max-w-6xl mx-auto">
    novoceo &middot; zigbee2mqtt monitoring
  </footer>

  <script>
    function applyTheme(mode) {
      var resolved = mode === 'system'
        ? (window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark')
        : mode;
      document.documentElement.classList.toggle('light', resolved === 'light');
      ['light', 'system', 'dark'].forEach(function(m) {
        var el = document.getElementById('btn-' + m);
        if (el) el.classList.toggle('active', m === mode);
      });
      localStorage.setItem('theme', mode);
    }
    function setTheme(mode) { applyTheme(mode); }
    (function(){
      var mode = localStorage.getItem('theme') || 'system';
      applyTheme(mode);
      window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', function() {
        if (localStorage.getItem('theme') === 'system') applyTheme('system');
      });
    })();
  </script>

</body>
</html>`
