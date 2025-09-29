// Copyright 2025 Aleksey Dobshikov
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

/*
Package router provides a centralized routing system for the git-sync service.
It manages all HTTP endpoints and provides a unified way to register and document routes.
*/
package router

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"git-sync/internal/config"
	"git-sync/logger"

	"github.com/justinas/alice"
)

// Route represents a registered HTTP route
type Route struct {
	Path        string
	Method      string
	Description string
	Handler     http.Handler
	HandlerFunc http.HandlerFunc
}

// Router manages all HTTP routes for the application
type Router struct {
	mu          sync.RWMutex
	routes      map[string]Route
	mux         *http.ServeMux
	middlewares alice.Chain
	config      *config.Config
}

// New creates a new Router instance
func New(cfg *config.Config) *Router {
	return &Router{
		routes: make(map[string]Route),
		mux:    http.NewServeMux(),
		config: cfg,
	}
}

// RegisterRoute registers a new route with the router
func (r *Router) RegisterRoute(path, method, description string, handler http.Handler, handlerFunc http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	logger.Debug("Registering route: %s %s - %s", method, path, description)

	route := Route{
		Path:        path,
		Method:      method,
		Description: description,
		Handler:     handler,
		HandlerFunc: handlerFunc,
	}

	r.routes[path] = route

	// Apply middleware to handlers
	if handler != nil {
		// Apply global middleware
		wrappedHandler := r.middlewares.Then(handler)
		r.mux.Handle(path, wrappedHandler)
	} else if handlerFunc != nil {
		// Apply global middleware
		wrappedHandler := r.middlewares.Then(handlerFunc)
		r.mux.Handle(path, wrappedHandler)
	}
}

// GetRoutes returns all registered routes
func (r *Router) GetRoutes() []Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var routes []Route
	for _, route := range r.routes {
		routes = append(routes, route)
	}

	// Sort routes by path
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	return routes
}

// GetMux returns the underlying ServeMux
func (r *Router) GetMux() *http.ServeMux {
	return r.mux
}

// SetupMiddleware configures the middleware chain based on config
func (r *Router) SetupMiddleware() {
	chain := alice.New()

	basicUsername := r.config.HttpServer.Auth.Username
	basicPassword := r.config.HttpServer.Auth.Password
	bearerToken := r.config.HttpServer.Auth.Token

	useBasicAuth := basicUsername != "" && basicPassword != ""
	useBearerToken := len(bearerToken) > 0

	// Add logging middleware for all requests
	chain = chain.Append(r.loggingMiddleware())

	switch {
	case useBasicAuth:
		chain = chain.Append(r.basicAuthMiddleware(basicUsername, basicPassword))
		if basicUsername == basicPassword && len(basicUsername) > 0 {
			logger.Warning("HTTP server: basic authentication (unsafe password)\n")
		} else {
			logger.Info("HTTP server: basic authentication\n")
			logger.Debug("Basic auth configured with username: %s", basicUsername)
		}
	case !useBasicAuth && useBearerToken:
		chain = chain.Append(r.bearerAuthMiddleware(bearerToken))
		logger.Info("HTTP server: token authentication\n")
		logger.Debug("Bearer token auth configured")
	default:
		logger.Info("HTTP server: no authentication\n")
		logger.Debug("No authentication configured for HTTP server")
	}

	r.middlewares = chain
}

// ApplyMiddleware applies the configured middleware to a handler
func (r *Router) ApplyMiddleware(handler http.Handler) http.Handler {
	// Check if middlewares chain is empty by checking if it has any constructors
	// Since alice.Chain doesn't expose its internal constructors, we'll just call Then
	// which will work correctly even with an empty chain
	return r.middlewares.Then(handler)
}

// StartServer starts the HTTP server with all registered routes
func (r *Router) StartServer(ctx context.Context) {
	addr := r.config.HttpServer.Addr

	if len(addr) == 0 {
		logger.Info("HTTP server: not started\n")
		return
	} else {
		logger.Info("HTTP server: http://%s", addr)
		logger.Debug("Starting HTTP server on address: %s", addr)
	}

	// Register the root handler to show all available routes
	r.RegisterRoute("/", "GET", "List of available endpoints", nil, r.rootHandler)

	go func() {
		server := &http.Server{
			Addr:    addr,
			Handler: r.mux,
		}

		logger.Debug("HTTP server listening on %s", addr)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			if logErr := logger.Error("HTTP server error: %v", err); logErr != nil {
				// Handle the error from logger.Error if needed
				fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
			}
			panic(err)
		}
	}()
}

// rootHandler displays an enhanced dashboard (vanilla HTML/CSS/JS)
func (r *Router) rootHandler(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get current year for footer
	currentYear := time.Now().Year()

	if _, err := fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8" />
<meta name="viewport" content="width=device-width,initial-scale=1" />
<title>Git Sync Dashboard</title>
<style>
:root{
  --bg:#f6f7f9; --card:#fff; --text:#0f172a; --muted:#6b7280; --border:#e5e7eb;
  --ok:#10b981; --warn:#f59e0b; --crit:#ef4444; --info:#3b82f6;
}
@media (prefers-color-scheme: dark){
  :root{ --bg:#0b0b0c; --card:#121316; --text:#e6e7eb; --muted:#a1a1aa; --border:#232428; }
}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:14px/1.45 ui-sans-serif,system-ui,-apple-system,"Segoe UI",Roboto,"Helvetica Neue",Arial}
.container{max-width:1100px;margin:0 auto;padding:20px}
.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:18px;padding:14px 18px;background:var(--card);border:1px solid var(--border);border-radius:12px}
.hleft{display:flex;align-items:center;gap:12px}
.logo{height:36px;width:36px;border-radius:9px;background:linear-gradient(135deg,#6366f1,#8b5cf6);display:flex;align-items:center;justify-content:center;color:#fff;font-weight:700}
.status{padding:6px 10px;border-radius:999px;color:#fff;font-weight:600}
.status.ok{background:var(--ok)} .status.warn{background:var(--warn)} .status.crit{background:var(--crit)}
.button{cursor:pointer;border:1px solid var(--border);background:var(--card);color:var(--text);padding:8px 12px;border-radius:10px}
.button.primary{background:var(--info);border-color:transparent;color:#fff}
.button[disabled]{opacity:.6;cursor:not-allowed}
.hero{background:var(--card);border:1px solid var(--border);border-radius:12px;padding:18px;margin-bottom:16px;text-align:center}
.kbd{border:1px solid var(--border);padding:2px 6px;border-radius:6px;font-family:ui-monospace,Menlo,monospace;font-size:12px}
.grid{display:grid;gap:12px}
.cards{grid-template-columns:repeat(auto-fit,minmax(260px,1fr))}
.card{background:var(--card);border:1px solid var(--border);border-left:4px solid var(--border);border-radius:12px;padding:12px}
.card.ok{border-left-color:var(--ok)} .card.warn{border-left-color:var(--warn)} .card.crit{border-left-color:var(--crit)}
.card h3{margin:0 0 6px 0;font-size:15px}
.small{color:var(--muted);font-size:12px}
.endpoints{background:var(--card);border:1px solid var(--border);border-radius:12px;padding:12px;margin-top:14px}
ul.ep{list-style:none;padding:0;margin:0}
ul.ep li{padding:8px 0;border-bottom:1px solid var(--border)}
ul.ep li:last-child{border-bottom:none}
a{color:var(--info);text-decoration:none}
a:hover{text-decoration:underline}
pre{background:#0b0b0c;color:#e6e7eb;padding:10px;border-radius:8px;overflow:auto;max-height:200px}
.badge{display:inline-flex;align-items:center;gap:6px;padding:2px 8px;border-radius:999px;font-size:12px}
.badge.ok{background:rgba(16,185,129,.15);color:var(--ok)}
.badge.warn{background:rgba(245,158,11,.15);color:var(--warn)}
.badge.crit{background:rgba(239,68,68,.15);color:var(--crit)}
.alert{display:flex;justify-content:space-between;align-items:center;border:1px solid var(--border);border-radius:10px;padding:8px 10px}
.alert + .alert{margin-top:8px}
.svgline{width:100%%;height:56px;display:block}
.footer{margin-top:18px;text-align:center;color:var(--muted);font-size:12px}
</style>
</head>
<body>
<div class="container">
  <header class="header">
    <div class="hleft">
      <div class="logo">GS</div>
      <div>
        <div style="font-weight:700">Git Sync Dashboard</div>
        <div class="small">Панель управления синхронизацией репозиториев</div>
      </div>
    </div>
    <div id="globalStatus" class="status ok">● Operational</div>
  </header>

  <section class="hero">
    <h2 style="margin:0 0 4px 0">Repository Sync Status</h2>
    <div class="small" id="lastSyncTime">Last Sync: —</div>
    <div class="small" id="lastSyncStatus"></div>
    <div style="margin-top:10px">
      <button id="btnRefresh" class="button" title="Обновить (R)">↻ Обновить</button>
      <button id="btnSync" class="button primary" title="Запустить синхронизацию (W)">🚀 Синхронизировать</button>
    </div>
    <div class="small" style="margin-top:8px">Горячие клавиши: <span class="kbd">R</span> — обновить, <span class="kbd">W</span> — синк</div>
  </section>

  <section class="grid cards">
    <div class="card" id="healthCard">
      <h3>Health</h3>
      <div id="healthSummary" class="small">—</div>
      <a href="/health" target="_blank">/health</a>
    </div>
    <div class="card" id="readyCard">
      <h3>Readiness</h3>
      <div id="readySummary" class="small">—</div>
      <a href="/ready" target="_blank">/ready</a>
    </div>
    <div class="card" id="versionCard">
      <h3>Version</h3>
      <div id="versionInfo" class="small">—</div>
      <a href="/version" target="_blank">/version</a>
    </div>
    <div class="card" id="metricsCard">
      <h3>Key Metrics</h3>
      <div id="metricsSummary" class="small">—</div>
      <div id="p95ChartWrap"><svg class="svgline" viewBox="0 0 100 20" preserveAspectRatio="none"><polyline id="p95Line" fill="none" stroke="currentColor" stroke-width="0.6" points=""/></svg></div>
      <a href="/metrics" target="_blank">/metrics</a>
    </div>
  </section>

  <section class="grid" style="margin-top:12px">
    <div class="card" id="alertsCard">
      <h3>Alerts</h3>
      <div id="alerts"></div>
    </div>
    <div class="card">
      <h3>Raw /status</h3>
      <pre id="statusRaw">—</pre>
    </div>
  </section>

  <section class="endpoints">
    <h3 style="margin:0 0 6px 0">Available Endpoints</h3>
    <ul class="ep">
`); err != nil {
		// Handle the error from fmt.Fprintf if needed
		if logErr := logger.Error("Error writing HTML response: %v", err); logErr != nil {
			// If we can't log the error, at least print it to stderr
			fmt.Fprintf(os.Stderr, "Error writing HTML response: %v\n", err)
		}
		return
	}

	// endpoints (как у тебя)
	var paths []string
	for path := range r.routes {
		if path != "/" {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	for _, path := range paths {
		if route, exists := r.routes[path]; exists {
			if _, err := fmt.Fprintf(w, `      <li><a href="%s">%s</a> — %s</li>`+"\n", path, path, route.Description); err != nil {
				// Handle the error from fmt.Fprintf if needed
				if logErr := logger.Error("Error writing route to response: %v", err); logErr != nil {
					// If we can't log the error, at least print it to stderr
					fmt.Fprintf(os.Stderr, "Error writing route to response: %v\n", err)
				}
				return
			}
		} else {
			if _, err := fmt.Fprintf(w, `      <li><a href="%s">%s</a></li>`+"\n", path, path); err != nil {
				// Handle the error from fmt.Fprintf if needed
				if logErr := logger.Error("Error writing route to response: %v", err); logErr != nil {
					// If we can't log the error, at least print it to stderr
					fmt.Fprintf(os.Stderr, "Error writing route to response: %v\n", err)
				}
				return
			}
		}
	}

	if _, err := fmt.Fprintf(w, `    </ul>
  </section>

  <footer class="footer">
    <div>© %d Git Sync Service</div>
  </footer>
</div>

<script>
// ---------- tiny utils ----------
const $ = (id)=>document.getElementById(id);
const sleep=(ms)=>new Promise(r=>setTimeout(r,ms));
const withTimeout = (p, ms=10000)=>Promise.race([p, new Promise((_,rej)=>setTimeout(()=>rej(new Error("timeout")),ms))]);

function setStatus(ok){
  const el=$("globalStatus");
  el.className = "status " + (ok?"ok":"crit");
  el.textContent = ok ? "● Operational" : "● Degraded";
}

// ---------- Prometheus text parser ----------
function parseProm(text){
  const lines = text.split(/\r?\n/);
  const samples = [];
  const byName = {};
  const histograms = {}; // base -> labelKey -> {buckets:[{le,value}],sum,count}
  for(const raw of lines){
    const line = raw.trim();
    if(!line || line.startsWith("#")) continue;
    const m = line.match(/^([a-zA-Z_:][a-zA-Z0-9_:]*)(?:\{([^}]*)\})?\s+(-?[0-9.]+(?:e[+-]?\d+)?)$/);
    if(!m) continue;
    const metric = m[1]; const lblStr = m[2]||""; const value = Number(m[3]);
    const labels = {};
    if(lblStr){
      let i=0;
      while(i<lblStr.length){
        const eq = lblStr.indexOf("=", i); if(eq<0) break;
        const key = lblStr.slice(i,eq).trim();
        let j = eq+1, v="";
        if(lblStr[j]==='"'){ j++; const start=j; for(;j<lblStr.length;j++){ if(lblStr[j]==='"' && lblStr[j-1] !== "\\") break; } v = lblStr.slice(start,j).replace(/\\"/g,'"'); j++; if(lblStr[j]===",") j++; }
        else { const c = lblStr.indexOf(",", j); v = c<0? lblStr.slice(j): lblStr.slice(j,c); j = c<0? lblStr.length: c+1; }
        labels[key]=v; i=j;
      }
    }
    const s = {metric, labels, value};
    samples.push(s); (byName[metric] ||= []).push(s);

    if (/_bucket$|_sum$|_count$/.test(metric)){
      const base = metric.replace(/_(bucket|sum|count)$/,"");
      const key = Object.keys(labels).filter(k=>k!=="le").sort().map(k=>k+"="+labels[k]).join(",");
      (histograms[base] ||= {}); (histograms[base][key] ||= {buckets:[]});
      if(metric.endsWith("_bucket")){
        const le = labels.le === "+Inf" ? Infinity : Number(labels.le);
        histograms[base][key].buckets.push({le,value});
      }else if(metric.endsWith("_sum")) histograms[base][key].sum=value;
      else if(metric.endsWith("_count")) histograms[base][key].count=value;
    }
  }
  // sort buckets
  for(const base of Object.keys(histograms)){
    for(const key of Object.keys(histograms[base])){
      histograms[base][key].buckets.sort((a,b)=>{
        const ax=a.le===Infinity?Number.POSITIVE_INFINITY:a.le;
        const bx=b.le===Infinity?Number.POSITIVE_INFINITY:b.le;
        return ax-bx;
      });
    }
  }
  return {samples, byName, histograms};
}
const getGauge=(byName,name,where)=> (byName[name]||[]).find(s=>!where||where(s));
function estimatePqFromHist(h,q){
  if(!h || !h.count || !h.buckets.length) return null;
  const target = h.count*q;
  let prevC=0, prevLe=0;
  for(const b of h.buckets){
    const c=b.value, upper=b.le;
    if(c>=target){
      const within=c-prevC; const pos = within<=0?0:(target-prevC)/within;
      const span = (upper===Infinity? prevLe*2: upper) - prevLe || (upper===Infinity? prevLe: upper);
      return prevLe + pos*span;
    }
    prevC=c; prevLe = (upper===Infinity? prevLe: upper);
  }
  return h.buckets.at(-1)?.le ?? null;
}

// ---------- fetchers ----------
async function json(url){ try{ const r = await withTimeout(fetch(url)); if(!r.ok) throw new Error(url+" -> "+r.status); return await r.json(); }catch(e){ return null; } }
async function text(url){ try{ const r = await withTimeout(fetch(url)); if(!r.ok) throw new Error(url+" -> "+r.status); return await r.text(); }catch(e){ return ""; } }

// ---------- UI update ----------
function fmtSec(s) { 
  if (!Number.isFinite(s) || s <= 0) return "—";
  if (s < 60) return Math.round(s) + "s";
  const m = Math.floor(s / 60);
  const r = Math.floor(s %% 60);
  return m + "m " + r + "s";
}
function cls(el, level){ 
  // Remove existing status classes
  el.classList.remove("ok", "warn", "crit");
  // Add the new status class if provided
  if (level) {
    el.classList.add(level);
  }
}

function putAlerts(items){
  const wrap = $("alerts"); wrap.innerHTML = "";
  if(!items.length){ wrap.innerHTML = '<div class="alert"><div class="badge ok">OK</div><div class="small">No alerts</div></div>'; return; }
  for(const a of items){
    const tone = a.sev==="crit" ? "crit" : "warn";
    const div = document.createElement("div");
    div.className="alert";
    div.innerHTML = '<div class="badge '+tone+'">'+a.sev.toUpperCase()+'</div><div class="small">'+a.label+' — '+a.tip+'</div>';
    wrap.appendChild(div);
  }
}

// New function to display rate limiting information
function putRateLimitInfo(parsed) {
  if (!parsed || !parsed.byName) return "";
  
  // Get rate limited IPs
  const rateLimitedMetrics = parsed.byName["webhook_rate_limited_ips"] || [];
  const requestsPerIPMetrics = parsed.byName["webhook_requests_per_ip"] || [];
  
  const rateLimitedIPs = rateLimitedMetrics
    .filter(s => s.value === 1)
    .map(s => s.labels.ip);
    
  const requestsPerIP = {};
  requestsPerIPMetrics.forEach(s => {
    requestsPerIP[s.labels.ip] = s.value;
  });
  
  if (rateLimitedIPs.length === 0 && Object.keys(requestsPerIP).length === 0) {
    return "Rate limiting: No active requests";
  }
  
  let info = "Rate limiting: ";
  if (Object.keys(requestsPerIP).length > 0) {
    const ipEntries = Object.entries(requestsPerIP)
      .map(([ip, count]) => ip+"("+count+")")
      .join(", ");
    info += ipEntries;
  }
  
  if (rateLimitedIPs.length > 0) {
    info += " | RATE LIMITED: " + rateLimitedIPs.join(", ");
  }
  
  return info;
}

function drawSpark(values){
  // values: array 0..1, draw polyline
  const pts = values.map((v,i)=> (i*(100/(values.length-1))).toFixed(2)+","+(20 - v*18).toFixed(2)).join(" ");
  $("p95Line").setAttribute("points", pts);
}

async function loadAll(){
  $("btnRefresh").disabled = true;
  try{
    const [health,status,version,ready,metricsRaw] = await Promise.all([
      json("/health"), json("/status"), json("/version"), json("/ready"), text("/metrics")
    ]);

    // status block
    $("statusRaw").textContent = status? JSON.stringify(status,null,2): "—";

    // version
    const v = version && (version.version||version.Version||"—");
    const c = version && (version.commit||version.Commit||"");
    $("versionInfo").textContent = "Version: "+v + (c? (" ("+String(c).slice(0,7)+")"):"");

    // health & readiness
    const parsed = metricsRaw? parseProm(metricsRaw): null;
    const healthy = Number(getGauge(parsed?.byName||{}, "service_healthy")?.value ?? 0) === 1;
    const readyOK = Number(getGauge(parsed?.byName||{}, "k8s_ready_status")?.value ?? 0) === 1;
    setStatus(healthy && readyOK);
    $("healthSummary").textContent = healthy? "Healthy": "Degraded";
    cls($("healthCard"), healthy? "ok" : "crit");
    $("readySummary").textContent = readyOK? "Ready": "Not Ready";
    cls($("readyCard"), readyOK? "ok" : "crit");

    // key metrics
    const syncErr = (parsed?.byName?.["sync_errors_total"]||[]).reduce((a,s)=>a+s.value,0);
    const syncReq = (parsed?.byName?.["sync_requests_total"]||[]).reduce((a,s)=>a+s.value,0);
    const driftMax = (parsed?.byName?.["repo_commit_drift"]||[]).reduce((a,s)=>Math.max(a,s.value),0);
    const lastAgeMax = (parsed?.byName?.["repo_last_sync_age_seconds"]||[]).reduce((a,s)=>Math.max(a,s.value),0);
    
    // Add rate limiting information to metrics summary
    const rateLimitInfo = putRateLimitInfo(parsed);
    $("metricsSummary").textContent = "Requests: "+syncReq+" | Errors: "+syncErr+" | Drift max: "+driftMax+" | LastSync age: "+fmtSec(lastAgeMax) + " | " + rateLimitInfo;

    // p95 sync_duration_seconds
    let p95 = null;
    const histMap = parsed?.histograms?.["sync_duration_seconds"] || {};
    const anyKey = Object.keys(histMap)[0];
    if(anyKey) p95 = estimatePqFromHist(histMap[anyKey], 0.95);
    // Draw tiny spark from buckets (normalized)
    if(anyKey){
      const buckets = histMap[anyKey].buckets;
      const maxV = Math.max(...buckets.map(b=>b.value||0)) || 1;
      const vals = buckets.slice(0,8).map(b=> (b.value||0)/maxV); // first 8 for small chart
      drawSpark(vals);
    }

    // alerts
    const alerts = [];
    if(syncErr > 0) alerts.push({sev:"crit", label:"sync_errors_total > 0", tip:"Проверьте error_type фильтрами"});
    if((p95||0) > 30) alerts.push({sev:"warn", label:"p95(sync_duration_seconds) > 30s", tip:"Оптимизируйте сеть/IO"});
    if(driftMax > 10) alerts.push({sev:"crit", label:"repo_commit_drift > 10", tip:"Учащайте синхронизацию"});
    if(lastAgeMax > 300) alerts.push({sev:"crit", label:"repo_last_sync_age_seconds > 300", tip:"Проверьте планировщик/вебхук"});
    if(!healthy) alerts.push({sev:"crit", label:"service_healthy == 0", tip:"Смотрите /health детали"});
    if(!readyOK) alerts.push({sev:"crit", label:"k8s_ready_status == 0", tip:"Readiness probe"});
    putAlerts(alerts);

    // last sync time/status (по данным метрик if available)
    const lastSuccess = (parsed?.byName?.["sync_status"]||[]).filter(s=>s.labels.metric==="last_success_timestamp").map(s=>s.value).sort().at(-1);
    if(lastSuccess) $("lastSyncTime").textContent = "Last Sync: "+ new Date(lastSuccess*1000).toLocaleString();
    else $("lastSyncTime").textContent = "Last Sync: —";
    $("lastSyncStatus").textContent = p95? ("p95: ~"+p95.toFixed(2)+"s"):"";

    // decorate metrics card
    cls($("metricsCard"), (syncErr>0 || driftMax>10 || lastAgeMax>300)? "warn" : "ok");
  }catch(e){
    console.error(e);
    setStatus(false);
    cls($("healthCard"), "crit"); cls($("readyCard"), "crit"); cls($("metricsCard"), "crit");
  }finally{
    $("btnRefresh").disabled = false;
  }
}

// ---------- sync trigger ----------
async function triggerSync(){
  const btn = $("btnSync"); const old = btn.textContent;
  btn.disabled = true; btn.textContent = "⏳ Syncing...";
  try{
    const r = await withTimeout(fetch("/webhook",{method:"POST"}), 15000);
    if(!r.ok) throw new Error("HTTP "+r.status);
    await sleep(1200);
    await loadAll();
  }catch(e){
    alert("Error triggering sync: "+e.message);
  }finally{
    btn.disabled = false; btn.textContent = old;
  }
}

// ---------- events ----------
$("btnRefresh").addEventListener("click", loadAll);
$("btnSync").addEventListener("click", triggerSync);
window.addEventListener("keydown", (e)=>{
  const k = e.key.toLowerCase();
  if(k==="r"){ e.preventDefault(); loadAll(); }
  if(k==="w"){ e.preventDefault(); triggerSync(); }
});

// initial + interval
loadAll();
setInterval(loadAll, 15000);
</script>
</body>
</html>`, currentYear); err != nil {
		// Handle the error from fmt.Fprintf if needed
		_, _ = fmt.Fprintf(os.Stderr, "Error writing HTML response: %v", err)
		return
	}
}

// basicAuthMiddleware provides basic authentication
func (r *Router) basicAuthMiddleware(username, password string) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Skip authentication for root endpoint (dashboard) only
			if req.URL.Path == "/" {
				next.ServeHTTP(w, req)
				return
			}

			// Skip authentication for webhook endpoint if configured to be enabled
			if req.URL.Path == "/webhook" && r.config.Webhook.Enabled {
				next.ServeHTTP(w, req)
				return
			}

			user, pass, ok := req.BasicAuth()
			if !ok || user != username || pass != password {
				logger.Debug("Basic auth failed for request from %s to %s", req.RemoteAddr, req.URL.Path)
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			logger.Debug("Basic auth successful for user %s from %s to %s", user, req.RemoteAddr, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	}
}

// bearerAuthMiddleware provides bearer token authentication
func (r *Router) bearerAuthMiddleware(token string) alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Skip authentication for root endpoint (dashboard) only
			if req.URL.Path == "/" {
				next.ServeHTTP(w, req)
				return
			}

			// Skip authentication for webhook endpoint if configured to be enabled
			if req.URL.Path == "/webhook" && r.config.Webhook.Enabled {
				next.ServeHTTP(w, req)
				return
			}

			authHeader := req.Header.Get("Authorization")
			if authHeader == "" {
				logger.Debug("Bearer auth failed - no Authorization header from %s to %s", req.RemoteAddr, req.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if authHeader != "Bearer "+token {
				logger.Debug("Bearer auth failed - invalid token from %s to %s", req.RemoteAddr, req.URL.Path)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			logger.Debug("Bearer auth successful from %s to %s", req.RemoteAddr, req.URL.Path)
			next.ServeHTTP(w, req)
		})
	}
}

// loggingMiddleware provides request logging
func (r *Router) loggingMiddleware() alice.Constructor {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()

			// Wrap the ResponseWriter to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, req)

			// Log the request
			duration := time.Since(start)
			logger.Debug("%s %s %d %v", req.Method, req.URL.Path, wrapped.statusCode, duration)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
