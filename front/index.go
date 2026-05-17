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
    .glow-on    { box-shadow: 0 0 30px rgba(52,211,153,.4),  0 0 80px rgba(52,211,153,.12);  border-color: rgba(52,211,153,.35)  !important; }
    .glow-off   { box-shadow: 0 0 20px rgba(139,92,246,.25), 0 0 60px rgba(139,92,246,.08);  border-color: rgba(139,92,246,.25)  !important; }
    .glow-alert { box-shadow: 0 0 30px rgba(239,68,68,.4),   0 0 80px rgba(239,68,68,.12);   border-color: rgba(239,68,68,.35)   !important; }
    .htmx-indicator { opacity: 0; transition: opacity .2s ease-in; }
    .htmx-request .htmx-indicator { opacity: 1; }
    .htmx-request.htmx-indicator  { opacity: 1; }
    @keyframes spin { to { transform: rotate(360deg); } }
    .spin { animation: spin .8s linear infinite; display: inline-block; }

    /* ── Light ── */
    html.light body              { background: #f1f5f9; }
    html.light .bg-slate-900     { background-color: #ffffff !important; }
    html.light .bg-slate-800     { background-color: #f1f5f9 !important; }
    html.light .text-slate-100   { color: #1e293b   !important; }
    html.light .text-slate-300   { color: #475569   !important; }
    html.light .text-slate-400   { color: #64748b   !important; }
    html.light .text-slate-500   { color: #64748b   !important; }
    html.light .text-slate-600   { color: #475569   !important; }
    html.light .text-slate-700   { color: #94a3b8   !important; }
    html.light .border-slate-700 { border-color: #e2e8f0 !important; }
    html.light .border-slate-800 { border-color: #f1f5f9 !important; }
    html.light .border-slate-900 { border-color: #e2e8f0 !important; }
    html.light .card-cyan   { box-shadow: 0 0 0 1px rgba(6,182,212,.25),  0 4px 24px rgba(6,182,212,.12); }
    html.light .card-amber  { box-shadow: 0 0 0 1px rgba(245,158,11,.25), 0 4px 24px rgba(245,158,11,.12); }
    html.light .card-emer   { box-shadow: 0 0 0 1px rgba(52,211,153,.25), 0 4px 24px rgba(52,211,153,.12); }
    html.light .card-violet { box-shadow: 0 0 0 1px rgba(139,92,246,.25), 0 4px 24px rgba(139,92,246,.12); }
    html.light .glow-on    { box-shadow: 0 0 20px rgba(52,211,153,.3),  0 0 50px rgba(52,211,153,.08);  border-color: rgba(52,211,153,.4)  !important; }
    html.light .glow-off   { box-shadow: 0 0 15px rgba(139,92,246,.2),  0 0 40px rgba(139,92,246,.06); border-color: rgba(139,92,246,.35) !important; }
    html.light .glow-alert { box-shadow: 0 0 20px rgba(239,68,68,.3),   0 0 50px rgba(239,68,68,.08);  border-color: rgba(239,68,68,.4)   !important; }
  </style>
  <script>
    (function(){
      if (localStorage.getItem('theme') === 'light')
        document.documentElement.classList.add('light');
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
    <div class="flex items-center gap-3">
      <button id="theme-btn" onclick="toggleTheme()"
        class="font-mono text-base text-slate-500 hover:text-slate-300 transition-colors leading-none select-none"
        title="Changer le thème">&#9728;</button>
      <span class="h-2 w-2 rounded-full bg-emerald-400 animate-pulse"></span>
    </div>
  </header>

  <main class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5 max-w-6xl mx-auto">

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
        <p class="font-mono text-xs text-slate-600">chargement...</p>
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
        <p class="font-mono text-xs text-slate-600">chargement...</p>
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
        <p class="font-mono text-xs text-slate-600">chargement...</p>
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
        <p class="font-mono text-xs text-slate-600">chargement...</p>
      </div>
    </div>

  </main>

  <footer class="mt-10 text-center font-mono text-xs text-slate-700 max-w-6xl mx-auto">
    novoceo &middot; zigbee2mqtt monitoring
  </footer>

  <script>
    function toggleTheme() {
      const light = document.documentElement.classList.toggle('light');
      localStorage.setItem('theme', light ? 'light' : 'dark');
      document.getElementById('theme-btn').innerHTML = light ? '&#9790;' : '&#9728;';
    }
    document.getElementById('theme-btn').innerHTML =
      document.documentElement.classList.contains('light') ? '&#9790;' : '&#9728;';
  </script>

</body>
</html>`
