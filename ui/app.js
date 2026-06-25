// ── State ───────────────────────────────────────────────────────────────────
const state = {
  ws: null,
  runs: [],
  activeRun: null,
  phases: [],
  agents: [],
  workflows: [],
  reviews: [],
}

const REVIEW_AFTER_PHASE = {
  'plan-review': 'planning',
  'task-review': 'architecture',
}

// ── Pixel art character sprites ──────────────────────────────────────────────
const PIXEL_CHARS = {
  odin:     `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='1' y='0' width='8' height='1' fill='#F59E0B'/><rect x='2' y='1' width='1' height='2' fill='#F59E0B'/><rect x='5' y='1' width='1' height='2' fill='#F59E0B'/><rect x='8' y='1' width='1' height='2' fill='#F59E0B'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#F59E0B'/><rect x='0' y='8' width='1' height='3' fill='#F59E0B'/><rect x='9' y='8' width='1' height='3' fill='#F59E0B'/><rect x='1' y='10' width='8' height='1' fill='#92400E'/><rect x='2' y='11' width='2' height='3' fill='#F59E0B'/><rect x='6' y='11' width='2' height='3' fill='#F59E0B'/><rect x='1' y='13' width='3' height='1' fill='#92400E'/><rect x='6' y='13' width='3' height='1' fill='#92400E'/></svg>`,
  tyrion:   `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='3' y='1' width='4' height='1' fill='#EF4444'/><rect x='2' y='2' width='6' height='4' fill='#FCD34D'/><rect x='3' y='3' width='1' height='1' fill='#0F172A'/><rect x='6' y='3' width='1' height='1' fill='#0F172A'/><rect x='4' y='5' width='2' height='1' fill='#92400E'/><rect x='1' y='6' width='8' height='5' fill='#EF4444'/><rect x='0' y='7' width='1' height='4' fill='#EF4444'/><rect x='9' y='7' width='2' height='3' fill='#FBBF24'/><rect x='1' y='10' width='8' height='1' fill='#B91C1C'/><rect x='2' y='11' width='2' height='3' fill='#EF4444'/><rect x='6' y='11' width='2' height='3' fill='#EF4444'/><rect x='1' y='13' width='3' height='1' fill='#B91C1C'/><rect x='6' y='13' width='3' height='1' fill='#B91C1C'/></svg>`,
  samwell:  `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='2' y='1' width='6' height='2' fill='#1E40AF'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='2' y='4' width='2' height='1' fill='#60A5FA'/><rect x='6' y='4' width='2' height='1' fill='#60A5FA'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='7' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='4' width='2' height='1' fill='#93C5FD'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#3B82F6'/><rect x='0' y='8' width='1' height='3' fill='#3B82F6'/><rect x='9' y='8' width='1' height='3' fill='#3B82F6'/><rect x='1' y='10' width='8' height='1' fill='#1D4ED8'/><rect x='2' y='11' width='2' height='3' fill='#3B82F6'/><rect x='6' y='11' width='2' height='3' fill='#3B82F6'/><rect x='1' y='13' width='3' height='1' fill='#1D4ED8'/><rect x='6' y='13' width='3' height='1' fill='#1D4ED8'/></svg>`,
  bran:     `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='4' y='0' width='2' height='3' fill='#8B5CF6'/><rect x='3' y='0' width='1' height='2' fill='#8B5CF6'/><rect x='6' y='0' width='1' height='2' fill='#8B5CF6'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#8B5CF6'/><rect x='0' y='8' width='1' height='3' fill='#8B5CF6'/><rect x='9' y='8' width='1' height='3' fill='#8B5CF6'/><rect x='1' y='10' width='8' height='1' fill='#6D28D9'/><rect x='2' y='11' width='2' height='3' fill='#8B5CF6'/><rect x='6' y='11' width='2' height='3' fill='#8B5CF6'/><rect x='1' y='13' width='3' height='1' fill='#6D28D9'/><rect x='6' y='13' width='3' height='1' fill='#6D28D9'/></svg>`,
  ragnar:   `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='0' y='1' width='1' height='2' fill='#F97316'/><rect x='9' y='1' width='1' height='2' fill='#F97316'/><rect x='1' y='2' width='8' height='2' fill='#F97316'/><rect x='2' y='4' width='6' height='4' fill='#FCD34D'/><rect x='3' y='5' width='1' height='1' fill='#0F172A'/><rect x='6' y='5' width='1' height='1' fill='#0F172A'/><rect x='4' y='7' width='2' height='1' fill='#92400E'/><rect x='1' y='8' width='8' height='3' fill='#F97316'/><rect x='0' y='8' width='1' height='3' fill='#F97316'/><rect x='9' y='8' width='1' height='3' fill='#F97316'/><rect x='1' y='10' width='8' height='1' fill='#C2410C'/><rect x='2' y='11' width='2' height='3' fill='#F97316'/><rect x='6' y='11' width='2' height='3' fill='#F97316'/><rect x='1' y='13' width='3' height='1' fill='#C2410C'/><rect x='6' y='13' width='3' height='1' fill='#C2410C'/></svg>`,
  lagertha: `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='2' y='0' width='6' height='1' fill='#EC4899'/><rect x='3' y='1' width='4' height='2' fill='#EC4899'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#EC4899'/><rect x='0' y='7' width='2' height='4' fill='#F9A8D4'/><rect x='9' y='8' width='1' height='3' fill='#EC4899'/><rect x='1' y='10' width='8' height='1' fill='#BE185D'/><rect x='2' y='11' width='2' height='3' fill='#EC4899'/><rect x='6' y='11' width='2' height='3' fill='#EC4899'/><rect x='1' y='13' width='3' height='1' fill='#BE185D'/><rect x='6' y='13' width='3' height='1' fill='#BE185D'/></svg>`,
  arya:     `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='4' y='0' width='2' height='1' fill='#0EA5E9'/><rect x='3' y='1' width='4' height='2' fill='#0F172A'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#0EA5E9'/><rect x='0' y='8' width='1' height='3' fill='#0EA5E9'/><rect x='9' y='7' width='2' height='4' fill='#BAE6FD'/><rect x='1' y='10' width='8' height='1' fill='#0369A1'/><rect x='2' y='11' width='2' height='3' fill='#0EA5E9'/><rect x='6' y='11' width='2' height='3' fill='#0EA5E9'/><rect x='1' y='13' width='3' height='1' fill='#0369A1'/><rect x='6' y='13' width='3' height='1' fill='#0369A1'/></svg>`,
  jonsnow:  `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='2' y='0' width='6' height='3' fill='#334155'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#CBD5E1'/><rect x='0' y='7' width='1' height='4' fill='#94A3B8'/><rect x='9' y='7' width='1' height='4' fill='#94A3B8'/><rect x='1' y='10' width='8' height='1' fill='#475569'/><rect x='2' y='11' width='2' height='3' fill='#94A3B8'/><rect x='6' y='11' width='2' height='3' fill='#94A3B8'/><rect x='1' y='13' width='3' height='1' fill='#334155'/><rect x='6' y='13' width='3' height='1' fill='#334155'/></svg>`,
  floki:    `<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 10 14'><rect x='1' y='0' width='1' height='3' fill='#06B6D4'/><rect x='8' y='0' width='1' height='3' fill='#06B6D4'/><rect x='1' y='2' width='8' height='1' fill='#06B6D4'/><rect x='2' y='3' width='6' height='4' fill='#FCD34D'/><rect x='3' y='4' width='1' height='1' fill='#0F172A'/><rect x='6' y='4' width='1' height='1' fill='#0F172A'/><rect x='4' y='6' width='2' height='1' fill='#92400E'/><rect x='1' y='7' width='8' height='4' fill='#06B6D4'/><rect x='0' y='8' width='1' height='3' fill='#06B6D4'/><rect x='9' y='8' width='2' height='2' fill='#A5F3FC'/><rect x='1' y='10' width='8' height='1' fill='#0891B2'/><rect x='2' y='11' width='2' height='3' fill='#06B6D4'/><rect x='6' y='11' width='2' height='3' fill='#06B6D4'/><rect x='1' y='13' width='3' height='1' fill='#0891B2'/><rect x='6' y='13' width='3' height='1' fill='#0891B2'/></svg>`,
}
const PIXEL_ART_URIS = {}
;(function() {
  Object.entries(PIXEL_CHARS).forEach(([k, svg]) => {
    PIXEL_ART_URIS[k] = 'data:image/svg+xml,' + encodeURIComponent(svg)
  })
})()
function getAvatarURI(agentName) {
  const p = getPersona(agentName)
  const key = (p ? p.display : agentName || '').toLowerCase().replace(/\s+/g, '')
  return PIXEL_ART_URIS[key] || null
}

// ── Agent personas — character display names + role icons ────────────────────
const AGENT_PERSONA = {
  'orchestrator':          { display: 'Odin',     icon: '⚡' },
  'coordinator':           { display: 'Odin',     icon: '⚡' },
  'planner':               { display: 'Tyrion',   icon: '◈'  },
  'requirements-analyst':  { display: 'Tyrion',   icon: '◈'  },
  'researcher':            { display: 'Samwell',  icon: '⊕'  },
  'tech-writer':           { display: 'Samwell',  icon: '⊕'  },
  'architect':             { display: 'Bran',     icon: '△'  },
  'senior-architect':      { display: 'Bran',     icon: '△'  },
  'api-designer':          { display: 'Bran',     icon: '△'  },
  'backend-eng':           { display: 'Ragnar',   icon: '⚙'  },
  'backend-engineer':      { display: 'Ragnar',   icon: '⚙'  },
  'frontend-eng':          { display: 'Lagertha', icon: '◱'  },
  'frontend-engineer':     { display: 'Lagertha', icon: '◱'  },
  'dba':                   { display: 'Ragnar',   icon: '⚙'  },
  'qa':                    { display: 'Arya',     icon: '✦'  },
  'qa-agent':              { display: 'Arya',     icon: '✦'  },
  'qa-engineer':           { display: 'Arya',     icon: '✦'  },
  'e2e-tester':            { display: 'Arya',     icon: '✦'  },
  'security':              { display: 'Jon Snow', icon: '⬡'  },
  'security-reviewer':     { display: 'Jon Snow', icon: '⬡'  },
  'code-reviewer':         { display: 'Jon Snow', icon: '⬡'  },
  'devops':                { display: 'Floki',    icon: '◎'  },
  'devops-agent':          { display: 'Floki',    icon: '◎'  },
  'devops-engineer':       { display: 'Floki',    icon: '◎'  },
}

// Phase-specific colors for the DAG
const PHASE_STYLE = {
  'planning':     { light: 'rgba(96,165,250,0.14)',  stroke: 'rgba(96,165,250,0.55)',  label: '#93C5FD', avatar: 'rgba(96,165,250,0.25)'  },
  'architecture': { light: 'rgba(167,139,250,0.14)', stroke: 'rgba(167,139,250,0.55)', label: '#C4B5FD', avatar: 'rgba(167,139,250,0.25)' },
  'engineering':  { light: 'rgba(245,158,11,0.14)',  stroke: 'rgba(245,158,11,0.55)',  label: '#FCD34D', avatar: 'rgba(245,158,11,0.25)'  },
  'qa':           { light: 'rgba(251,146,60,0.14)',  stroke: 'rgba(251,146,60,0.55)',  label: '#FDBA74', avatar: 'rgba(251,146,60,0.25)'  },
  'devops':       { light: 'rgba(34,197,94,0.14)',   stroke: 'rgba(34,197,94,0.55)',   label: '#86EFAC', avatar: 'rgba(34,197,94,0.25)'   },
  'security':     { light: 'rgba(239,68,68,0.14)',   stroke: 'rgba(239,68,68,0.55)',   label: '#FCA5A5', avatar: 'rgba(239,68,68,0.25)'   },
}

function getPersona(agentName) {
  const key = (agentName || '').toLowerCase().replace(/_/g, '-').trim()
  if (AGENT_PERSONA[key]) return AGENT_PERSONA[key]
  // fuzzy match: check if any key is a substring
  for (const [k, v] of Object.entries(AGENT_PERSONA)) {
    if (key.includes(k) || k.includes(key)) return v
  }
  const words = (agentName || '').split(/[-_ ]+/)
  const icon = words[0] ? words[0][0].toUpperCase() : '◈'
  return { display: agentName, icon }
}

function getPhaseStyle(phaseId) {
  return PHASE_STYLE[phaseId] || {
    light: 'rgba(71,85,105,0.14)', stroke: 'rgba(71,85,105,0.55)',
    label: '#64748B', avatar: 'rgba(71,85,105,0.25)'
  }
}

// ── Markdown renderer ────────────────────────────────────────────────────────
function renderMarkdown(raw) {
  if (!raw) return ''
  let s = raw
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;')

  // Fenced code blocks
  s = s.replace(/```[\w]*\n([\s\S]*?)```/g, (_,c) =>
    `<pre><code>${c.trimEnd()}</code></pre>`)

  // Inline code
  s = s.replace(/`([^`]+)`/g, '<code>$1</code>')

  // Headings
  s = s.replace(/^### (.+)$/gm, '<h3>$1</h3>')
  s = s.replace(/^## (.+)$/gm, '<h2>$1</h2>')
  s = s.replace(/^# (.+)$/gm, '<h1>$1</h1>')

  // HR
  s = s.replace(/^---$/gm, '<hr>')

  // Bold
  s = s.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')

  // Italic
  s = s.replace(/\*([^*]+)\*/g, '<em>$1</em>')

  // Links
  s = s.replace(/\[([^\]]+)\]\((https?:\/\/[^\)]+)\)/g,
    '<a href="$2" target="_blank" rel="noopener">$1</a>')

  // Tables (simple: |col|col|)
  s = s.replace(/((?:\|.+\|\n)+)/g, tbl => {
    const rows = tbl.trim().split('\n')
    let html = '<table>'
    rows.forEach((row, i) => {
      if (/^\|[-:| ]+\|$/.test(row)) return
      const cells = row.split('|').slice(1,-1).map(c => c.trim())
      const tag = i === 0 ? 'th' : 'td'
      html += '<tr>' + cells.map(c => `<${tag}>${c}</${tag}>`).join('') + '</tr>'
    })
    return html + '</table>'
  })

  // Unordered list
  s = s.replace(/((?:^[ \t]*[-*] .+\n?)+)/gm, block => {
    const items = block.trim().split('\n')
      .map(l => `<li>${l.replace(/^[ \t]*[-*] /,'')}</li>`).join('')
    return `<ul>${items}</ul>`
  })

  // Ordered list
  s = s.replace(/((?:^\d+\. .+\n?)+)/gm, block => {
    const items = block.trim().split('\n')
      .map(l => `<li>${l.replace(/^\d+\. /,'')}</li>`).join('')
    return `<ol>${items}</ol>`
  })

  // Paragraphs (double newline → <p>)
  s = s.split(/\n\n+/).map(para => {
    para = para.trim()
    if (!para) return ''
    if (/^<(h[1-3]|ul|ol|pre|table|hr)/.test(para)) return para
    return `<p>${para.replace(/\n/g,' ')}</p>`
  }).join('\n')

  return s
}

function updateOnboardingVisibility() {
  const hasRuns = state.runs && state.runs.length > 0
  const ob = document.getElementById('onboarding-card')
  const rd = document.getElementById('run-detail')
  const inspector = document.getElementById('inspector')
  if (!ob || !rd) return
  ob.style.display = hasRuns ? 'none' : 'flex'
  rd.style.display = hasRuns ? 'block' : 'none'
  if (inspector) inspector.style.display = (hasRuns && state.activeRun) ? 'flex' : 'none'
}

function toggleSection(bodyId) {
  const body = document.getElementById(bodyId)
  if (!body) return
  const isOpen = body.classList.contains('open')
  body.classList.toggle('open', !isOpen)
  body.style.display = isOpen ? 'none' : 'block'
  // Toggle chevron
  const chevronId = bodyId.replace('-body', '-chevron')
  const chevron = document.getElementById(chevronId)
  if (chevron) chevron.classList.toggle('open', !isOpen)
  // Toggle header border
  const header = body.previousElementSibling
  if (header) header.classList.toggle('open', !isOpen)
}

// (legacy constants removed — replaced by new renderTreeSimple layout)

// ── Inspector tab switching ──────────────────────────────────────────────────
function switchInspectorTab(tab) {
  document.querySelectorAll('.itab').forEach(b => b.classList.remove('active'))
  document.querySelectorAll('.itab-panel').forEach(p => p.style.display = 'none')
  const btn = document.getElementById(`itab-btn-${tab}`)
  const panel = document.getElementById(`itab-${tab}`)
  if (btn) btn.classList.add('active')
  if (panel) panel.style.display = 'flex'
}

// ── Boot ────────────────────────────────────────────────────────────────────
async function init() {
  await loadWorkflows()
  await loadRuns()
  await loadStats()
  connectWS()
  renderTreeSimple()
  updateOnboardingVisibility()
  switchInspectorTab('agent')
  // Poll active run every 6s — agent results land in SQLite async
  setInterval(async () => {
    if (state.activeRun) await refreshActiveRun()
  }, 6000)
}

async function refreshActiveRun() {
  try {
    const res = await fetch(`/api/runs/${state.activeRun.id}`)
    if (!res.ok) return
    const detail = await res.json()
    const prevCount = state.agents.length
    state.activeRun = detail
    state.phases = detail.phases || []
    state.agents = (detail.results || []).map(mapResult)
    state.reviews = (detail.reviews || []).map(r => ({
      gate: r.gate,
      status: r.status,
      summary: r.summary || '',
    }))
    renderReviewBanner()
    renderPhaseBar()
    if (state.agents.length !== prevCount) {
      renderTreeSimple()
      if (state.agents.length > 0) selectAgentData(state.agents[state.agents.length - 1])
    }
    await renderPhaseOutputs(state.activeRun.id)
    await renderOrchestrationCard(state.activeRun.id)
    // also refresh run list to catch status changes
    const rres = await fetch('/api/runs')
    if (rres.ok) { state.runs = await rres.json() || []; renderRunHistory(); updateOnboardingVisibility() }
  } catch (_) {}
}

// ── WebSocket ────────────────────────────────────────────────────────────────
function connectWS() {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  state.ws = new WebSocket(`${proto}//${location.host}/ws`)
  state.ws.onopen = () => {
    document.getElementById('ws-dot').className = 'ws-dot connected'
  }
  state.ws.onclose = () => {
    document.getElementById('ws-dot').className = 'ws-dot error'
    setTimeout(connectWS, 3000)
  }
  state.ws.onmessage = (e) => {
    try {
      const evt = JSON.parse(e.data)
      if (evt.type === 'agent_result') { onAgentResult(evt.payload); loadStats() }
      if (evt.type === 'review_pending') { onReviewPending(evt.payload) }
      if (evt.type === 'review_resolved') { onReviewResolved(evt.payload) }
    } catch (_) {}
  }
}

async function loadStats() {
  try {
    const res = await fetch('/api/stats')
    if (!res.ok) return
    const s = await res.json()
    renderStats(s)
  } catch (_) {}
}

function renderStats(s) {
  const panel = document.getElementById('sidebar-stats')
  if (!panel) return
  if (!s || s.runs_total === 0) { panel.style.display = 'none'; return }
  panel.style.display = 'block'
  document.getElementById('stat-agents-line').textContent =
    `${fmtNum(s.agents_total)} agents dispatched`
  document.getElementById('stat-tokens-line').textContent =
    `${fmtNum(s.tokens_total)} tokens · ${s.parallelism_speedup.toFixed(1)}× speedup`
}

function fmtNum(n) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M'
  if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
  return String(n)
}

function onAgentResult(result) {
  if (!result) return
  const mapped = mapResult(result)
  const idx = state.agents.findIndex(a => a.agent === mapped.agent && a.phase === mapped.phase && a.run_id === mapped.run_id)
  if (idx >= 0) {
    state.agents[idx] = mapped
  } else {
    state.agents.push(mapped)
  }
  renderTreeSimple()
  selectAgentData(mapped)
  renderPhaseBar()
}

function mapResult(r) {
  return {
    run_id: r.run_id,
    phase: r.phase_id || 'unknown',
    agent: r.agent,
    status: (r.status || 'pending').toLowerCase().replace(/ /g, '_'),
    conf: r.confidence || '',
    tokens: r.tokens_used || 0,
    summary: r.summary || '',
    sources: (() => { try { return typeof r.sources === 'string' ? JSON.parse(r.sources) : (r.sources || []) } catch(_){ return [] } })(),
    deliverables: (() => { try { return typeof r.deliverables === 'string' ? JSON.parse(r.deliverables) : (r.deliverables || []) } catch(_){ return [] } })(),
  }
}

function onReviewPending(payload) {
  if (!state.activeRun || state.activeRun.id !== payload.run_id) return
  const idx = state.reviews.findIndex(r => r.gate === payload.gate)
  const item = { gate: payload.gate, status: 'pending', summary: payload.summary || '' }
  if (idx >= 0) state.reviews[idx] = item
  else state.reviews.push(item)
  renderReviewBanner()
  renderPhaseBar()
}

function onReviewResolved(payload) {
  if (!state.activeRun || state.activeRun.id !== payload.run_id) return
  const r = state.reviews.find(r => r.gate === payload.gate)
  if (r) r.status = payload.status
  renderReviewBanner()
  renderPhaseBar()
}

function renderReviewBanner() {
  const container = document.getElementById('review-banner-container')
  if (!container) return
  const pending = state.reviews.find(r => r.status === 'pending')
  if (!pending) {
    container.style.display = 'none'
    container.innerHTML = ''
    return
  }
  const label = pending.gate === 'plan-review' ? 'PLAN REVIEW' : 'TASK REVIEW'
  container.style.display = 'block'
  container.innerHTML = `
    <div class="review-banner">
      <span class="review-banner-icon">&#9208;</span>
      <div class="review-banner-body">
        <span class="review-banner-title">${label} — Action required in Claude Code terminal</span>
        ${pending.summary ? `<p class="review-banner-summary">${esc(pending.summary)}</p>` : ''}
      </div>
    </div>
  `
}

// ── Data ─────────────────────────────────────────────────────────────────────
async function loadWorkflows() {
  try {
    const res = await fetch('/api/workflows')
    state.workflows = await res.json() || []
  } catch (_) { state.workflows = [] }
}

async function loadRuns() {
  try {
    const res = await fetch('/api/runs')
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    state.runs = await res.json() || []
    document.getElementById('offline-banner').style.display = 'none'
  } catch (_) {
    state.runs = []
    document.getElementById('offline-banner').style.display = 'block'
  }
  renderRunHistory()
  updateOnboardingVisibility()
  if (state.runs.length > 0) await loadRunDetail(state.runs[0].id)
}

async function loadRunDetail(runId) {
  // mark active in sidebar
  document.querySelectorAll('.run-item').forEach(el => el.classList.remove('active'))
  const active = document.querySelector(`.run-item[data-id="${runId}"]`)
  if (active) active.classList.add('active')

  try {
    const res = await fetch(`/api/runs/${runId}`)
    if (!res.ok) return
    const detail = await res.json()
    state.activeRun = detail
    state.phases = detail.phases || []
    state.agents = (detail.results || []).map(mapResult)
    state.reviews = (detail.reviews || []).map(r => ({
      gate: r.gate,
      status: r.status,
      summary: r.summary || '',
    }))
    renderReviewBanner()
    renderPhaseBar()
    renderTreeSimple()
    if (state.agents.length > 0) selectAgentData(state.agents[state.agents.length - 1])
    else document.getElementById('active-card').style.display = 'none'

    // Update header run title
    const titleEl = document.getElementById('hdr-run-title')
    if (titleEl) {
      titleEl.textContent = `${detail.id} · ${detail.workflow_name}`
      titleEl.style.display = 'inline'
    }

    // Show inspector
    document.getElementById('inspector').style.display = 'flex'

    // New: load orchestration card and phase outputs
    await renderOrchestrationCard(runId)
    await renderPhaseOutputs(runId)
  } catch (_) {}
}

// ── Renders ──────────────────────────────────────────────────────────────────
function renderRunHistory() {
  const el = document.getElementById('run-history')
  if (!state.runs.length) {
    el.innerHTML = '<div style="color:var(--muted);font-size:10px;padding:8px 10px">No runs yet</div>'
    return
  }
  el.innerHTML = state.runs.map(r => {
    const isRunning = r.status === 'running'
    const dotColor = isRunning ? 'var(--amber)' :
                     r.status === 'done'    ? 'var(--green)' :
                     r.status === 'blocked' ? 'var(--red)'   : 'var(--muted)'
    const dotGlow  = isRunning ? ';box-shadow:0 0 5px rgba(245,158,11,0.55)' : ''
    const isActive = state.activeRun && r.id === state.activeRun.id
    const taskExcerpt = (r.task_text || '').slice(0, 40) + ((r.task_text || '').length > 40 ? '…' : '')
    const showResume = isRunning && (Math.floor(Date.now() / 1000) - r.started_at) > 60
    const resumeBtn = showResume
      ? `<button class="ri-resume-btn" onclick="event.stopPropagation();openResumeModal('${esc(r.id)}')">Resume</button>`
      : ''
    return `
      <div class="run-item${isActive ? ' active' : ''}" data-id="${esc(r.id)}" onclick="loadRunDetail('${esc(r.id)}')">
        <div class="ri-header">
          <span class="status-badge" style="background:${dotColor}${dotGlow}"></span>
          <span class="ri-name">${esc(r.workflow_name || r.id)}</span>
          ${resumeBtn}
        </div>
        <div class="ri-meta">${esc(taskExcerpt || r.id)} · ${fmtTime(r.started_at)}</div>
      </div>`
  }).join('')
}

function openResumeModal(runId) {
  const existing = document.getElementById('resume-modal')
  if (existing) existing.remove()

  const cmd = `/team-resume ${runId}`
  const overlay = document.createElement('div')
  overlay.id = 'resume-modal'
  overlay.className = 'resume-modal-overlay'
  overlay.innerHTML = `
    <div class="resume-modal">
      <div class="resume-modal-title">Resume Run</div>
      <div class="resume-modal-body">
        <p class="resume-modal-hint">Paste this command into your Claude Code session to resume the run:</p>
        <div class="resume-modal-cmd">${esc(cmd)}</div>
        <div class="resume-modal-actions">
          <button class="resume-copy-btn" onclick="copyResumeCmd('${esc(cmd)}', this)">Copy command</button>
          <button class="resume-close-btn" onclick="document.getElementById('resume-modal').remove()">Close</button>
        </div>
      </div>
    </div>`
  overlay.addEventListener('click', e => {
    if (e.target === overlay) overlay.remove()
  })
  document.body.appendChild(overlay)
}

function copyResumeCmd(cmd, btn) {
  navigator.clipboard.writeText(cmd).then(() => {
    const orig = btn.textContent
    btn.textContent = 'Copied!'
    setTimeout(() => { btn.textContent = orig }, 1800)
  })
}

function renderPhaseBar() {
  const bar = document.getElementById('phase-bar')
  if (!state.phases.length) { bar.innerHTML = ''; return }

  const gateByPhase = {}
  Object.entries(REVIEW_AFTER_PHASE).forEach(([gate, phase]) => {
    // reviews ordered id ASC — last entry per gate is current state (handles reject-reloop)
    const reviews = state.reviews.filter(r => r.gate === gate)
    if (reviews.length) gateByPhase[phase] = reviews[reviews.length - 1]
  })

  bar.innerHTML = state.phases.map((p, i) => {
    let html = i > 0 ? '<span class="phase-arrow">→</span>' : ''
    html += `<span class="phase-pill ${esc(p.status || 'pending')}">${esc(p.phase_id)}</span>`
    const review = gateByPhase[p.phase_id]
    if (review) {
      const cls = review.status === 'pending'  ? 'review-chip--pending'  :
                  review.status === 'approved' ? 'review-chip--approved' : 'review-chip--rejected'
      const lbl = review.status === 'pending'  ? '&#9208; Awaiting Review' :
                  review.status === 'approved' ? '&#10003; Approved'      : '&#10007; Rejected'
      html += `<span class="phase-arrow">→</span><span class="review-chip ${cls}">${lbl}</span>`
    }
    return html
  }).join('')
}

function phaseOrder(p) {
  const idx = state.phases.findIndex(ph => ph.phase_id === p)
  return idx >= 0 ? idx : 999
}

function renderTreeSimple() {
  const svg = document.getElementById('tree-svg')
  if (!svg) return

  const W = svg.parentElement ? Math.max(svg.parentElement.clientWidth || 700, 400) : 700

  // ── Empty / loading states ────────────────────────────────────────────────
  if (!state.agents.length) {
    const isRunning = state.activeRun && (state.activeRun.status === 'running' || state.activeRun.status === 'pending')
    if (state.runs.length === 0) return

    svg.setAttribute('viewBox', `0 0 ${W} 100`)
    svg.removeAttribute('height')

    if (isRunning) {
      const isPending = state.activeRun.status === 'pending'
      const msg = isPending ? 'Waiting for /team-dispatch in Claude Code…' : esc(state.activeRun.id) + ' — agents working…'
      const sub = isPending ? 'Run /team-dispatch in your Claude session' : 'Diagram updates when first agent reports'
      const fill = isPending ? '#38BDF8' : '#F59E0B'
      svg.innerHTML = `
        <text x="${W/2}" y="40" text-anchor="middle" style="fill:${fill};font-family:system-ui,sans-serif;font-size:13px;font-weight:600">${msg}</text>
        <text x="${W/2}" y="60" text-anchor="middle" style="fill:#64748B;font-family:system-ui,sans-serif;font-size:11px">${sub}</text>`
    } else {
      svg.innerHTML = `
        <text x="${W/2}" y="48" text-anchor="middle" style="fill:#475569;font-family:system-ui,sans-serif;font-size:12px">Select a run to view the agent flow diagram</text>`
    }
    return
  }

  // ── Layout constants ──────────────────────────────────────────────────────
  const PH_W = 192, PH_H = 58    // phase node size
  const AG_W = 156, AG_H = 92    // agent chip size (narrower to reduce overlap)
  const AG_COL_GAP = 10          // gap between agent columns
  const AG_ROW_GAP = 14          // gap between agent rows
  const START_X = 24
  const PHASE_Y = 16
  const AGENT_START_Y = PHASE_Y + PH_H + 32  // gap between phase node and agent chips
  const MIN_PH_GAP = 28          // min gap between adjacent phase footprints

  // ── Group agents by phase ─────────────────────────────────────────────────
  const phaseMap = {}
  // Pre-populate all pipeline phases including pending ones
  if (state.phases && state.phases.length) {
    state.phases.forEach(p => { if (!phaseMap[p.phase_id]) phaseMap[p.phase_id] = [] })
  }
  state.agents.forEach(a => {
    if (!phaseMap[a.phase]) phaseMap[a.phase] = []
    phaseMap[a.phase].push(a)
  })
  const phases = Object.keys(phaseMap).sort((a, b) => phaseOrder(a) - phaseOrder(b))

  function phaseStatus(ph) {
    const sp = state.phases.find(p => p.phase_id === ph)
    if (sp) return sp.status || 'pending'
    const ag = phaseMap[ph] || []
    if (ag.some(a => a.status === 'blocked')) return 'blocked'
    if (ag.some(a => a.status === 'running')) return 'running'
    if (ag.every(a => a.status === 'done' || a.status === 'done_with_concerns')) return 'done'
    return 'pending'
  }

  // ── Compute phase center-x (footprint-aware to prevent agent chip overlap) ──
  function phaseFootprintHalf(ph) {
    const { cols } = agentGridDims(ph)
    const agW = cols * AG_W + (cols - 1) * AG_COL_GAP
    return Math.max(PH_W, agW) / 2 + 8
  }

  const phaseX = {}
  let cursorX = START_X
  phases.forEach((ph, i) => {
    cursorX += phaseFootprintHalf(ph)
    phaseX[ph] = cursorX
    cursorX += phaseFootprintHalf(ph)
    if (i + 1 < phases.length) cursorX += MIN_PH_GAP
  })
  const totalW = cursorX + START_X

  // ── Compute agent layout height per phase ─────────────────────────────────
  function agentGridDims(ph) {
    const count = phaseMap[ph].length
    const cols = count > 3 ? 2 : 1
    const rows = Math.ceil(count / cols)
    return { cols, rows }
  }
  function agentGridHeight(ph) {
    const { rows } = agentGridDims(ph)
    return rows * AG_H + (rows - 1) * AG_ROW_GAP
  }

  const maxAgentGridH = Math.max(...phases.map(ph => agentGridHeight(ph)), 0)
  const totalH = AGENT_START_Y + maxAgentGridH + 20
  // totalW computed above in footprint-aware phase layout

  svg.setAttribute('viewBox', `0 0 ${totalW} ${totalH}`)
  svg.setAttribute('width', String(totalW))
  svg.setAttribute('height', String(Math.max(totalH, 280)))

  // ── Build SVG ─────────────────────────────────────────────────────────────
  let defs = '<defs>'

  // Glow filter for running nodes
  defs += `<filter id="glow-run" x="-30%" y="-30%" width="160%" height="160%">
    <feGaussianBlur stdDeviation="4" result="blur"/>
    <feMerge><feMergeNode in="blur"/><feMergeNode in="SourceGraphic"/></feMerge>
  </filter>`

  // Arrowhead markers per edge index
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const dst = phaseStatus(ph)
    const arrowFill = dst === 'done' ? 'rgba(34,197,94,0.6)'
                    : dst === 'running' ? 'rgba(56,189,248,0.6)'
                    : 'rgba(71,85,105,0.5)'
    defs += `<marker id="arr-${pi}" markerWidth="7" markerHeight="7" refX="6" refY="3.5" orient="auto">
      <path d="M0,0.5 L0,6.5 L7,3.5 z" fill="${arrowFill}"/>
    </marker>`
  })

  // Hidden signal paths for orb animations
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    const x1 = phaseX[prevPh] + PH_W / 2
    const x2 = phaseX[ph] - PH_W / 2
    const y  = PHASE_Y + PH_H / 2
    const mx = (x1 + x2) / 2
    defs += `<path id="sp-${pi}" d="M${x1},${y} C${mx},${y} ${mx},${y} ${x2},${y}" fill="none"/>`
  })

  // Avatar clip paths (objectBoundingBox so circle clips correctly regardless of position)
  phases.forEach((ph) => {
    phaseMap[ph].forEach((ag, ai) => {
      defs += `<clipPath id="ac-${esc(ph)}-${ai}" clipPathUnits="objectBoundingBox"><circle cx=".5" cy=".5" r=".5"/></clipPath>`
    })
  })

  defs += '</defs>'

  let body = ''

  // ── Phase → next phase edges ──────────────────────────────────────────────
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    const x1 = phaseX[prevPh] + PH_W / 2
    const x2 = phaseX[ph] - PH_W / 2
    const y  = PHASE_Y + PH_H / 2
    const mx = (x1 + x2) / 2
    const dst = phaseStatus(ph)
    const edgeColor = dst === 'done'    ? 'rgba(34,197,94,0.5)'
                    : dst === 'running' ? 'rgba(56,189,248,0.45)'
                    : 'rgba(51,65,85,0.6)'
    const dashArray = dst === 'pending' ? '5 4' : 'none'
    const marchAnim = dst === 'running'
      ? `<animate attributeName="stroke-dashoffset" from="0" to="-18" dur="1.3s" repeatCount="indefinite"/>`
      : ''
    body += `<path d="M${x1},${y} C${mx},${y} ${mx},${y} ${x2},${y}"
      fill="none" stroke="${edgeColor}" stroke-width="1.5"
      stroke-dasharray="${dst === 'running' ? '5 4' : dashArray}"
      marker-end="url(#arr-${pi})">${marchAnim}</path>`
  })

  // ── Signal orbs ───────────────────────────────────────────────────────────
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    if (phaseStatus(ph) !== 'running' && phaseStatus(prevPh) !== 'running') return
    const dur = (1.6 + pi * 0.2).toFixed(2)
    const delay = ((pi - 1) * 0.35).toFixed(2)
    // Lead orb
    body += `<circle r="4.5" fill="#38BDF8" opacity="0.85">
      <animateMotion dur="${dur}s" repeatCount="indefinite" begin="${delay}s"><mpath href="#sp-${pi}"/></animateMotion>
    </circle>`
    // Trailing orb (smaller, offset)
    const trailDelay = (parseFloat(delay) + parseFloat(dur) * 0.45).toFixed(2)
    body += `<circle r="2.5" fill="#7DD3FC" opacity="0.55">
      <animateMotion dur="${dur}s" repeatCount="indefinite" begin="${trailDelay}s"><mpath href="#sp-${pi}"/></animateMotion>
    </circle>`
  })

  // ── Phase nodes ───────────────────────────────────────────────────────────
  phases.forEach((ph, pi) => {
    const cx = phaseX[ph]
    const nx = cx - PH_W / 2
    const ny = PHASE_Y
    const st = phaseStatus(ph)
    const ps = getPhaseStyle(ph)

    // Status-override colors when running/done/blocked
    let nodeFill   = ps.light
    let nodeStroke = ps.stroke
    let labelFill  = ps.label

    let strokeW = 1.5
    if (st === 'done')    { nodeFill = 'rgba(34,197,94,0.10)';  nodeStroke = 'rgba(34,197,94,0.5)';  labelFill = '#86EFAC' }
    if (st === 'running') { nodeFill = 'rgba(245,158,11,0.20)'; nodeStroke = '#F59E0B';               labelFill = '#FDE68A'; strokeW = 2.5 }
    if (st === 'blocked') { nodeFill = 'rgba(239,68,68,0.10)';  nodeStroke = 'rgba(239,68,68,0.5)';  labelFill = '#FCA5A5' }

    const pulseAnim = st === 'running'
      ? `<animate attributeName="stroke-width" values="${strokeW};${strokeW + 1.5};${strokeW}" dur="1.6s" repeatCount="indefinite"/>
         <animate attributeName="opacity" values="1;0.82;1" dur="1.6s" repeatCount="indefinite"/>`
      : ''

    const statusIcon = st === 'done' ? '✓' : st === 'running' ? '▸' : st === 'blocked' ? '✗' : '○'
    const statusColor = st === 'done' ? '#22C55E' : st === 'running' ? '#F59E0B' : st === 'blocked' ? '#EF4444' : '#475569'
    const statusText = st === 'done' ? 'complete' : st === 'running' ? 'in progress' : st === 'blocked' ? 'blocked' : 'pending'

    body += `
      ${st === 'running' ? `<rect x="${nx - 4}" y="${ny - 4}" width="${PH_W + 8}" height="${PH_H + 8}" rx="15" fill="rgba(245,158,11,0.07)" stroke="rgba(245,158,11,0.18)" stroke-width="1"/>` : ''}
      <rect x="${nx}" y="${ny}" width="${PH_W}" height="${PH_H}" rx="12"
        fill="${nodeFill}" stroke="${nodeStroke}" stroke-width="${strokeW}"
        ${st === 'running' ? 'filter="url(#glow-run)"' : ''}>${pulseAnim}</rect>
      <text x="${cx}" y="${ny + 24}" text-anchor="middle"
        style="fill:${labelFill};font-family:system-ui,sans-serif;font-size:11px;font-weight:700;letter-spacing:0.08em;text-transform:uppercase">
        ${esc(ph)}</text>
      <text x="${cx}" y="${ny + 41}" text-anchor="middle"
        style="fill:${statusColor};font-family:system-ui,sans-serif;font-size:9px;font-weight:600;letter-spacing:0.04em">
        ${statusIcon} ${statusText}</text>`

    // Drop line + fan connectors from phase to each agent chip
    const dropY1 = ny + PH_H
    const dropMid = AGENT_START_Y - 10
    if (phaseMap[ph].length > 0) {
      const agCount = phaseMap[ph].length
      const { cols: agCols } = agentGridDims(ph)
      const totalWchips = agCols * AG_W + (agCols - 1) * AG_COL_GAP
      const gsx = cx - totalWchips / 2
      // Vertical stem
      body += `<line x1="${cx}" y1="${dropY1}" x2="${cx}" y2="${dropMid}"
        stroke="${nodeStroke}" stroke-width="1" stroke-dasharray="3 2" opacity="0.35"/>`
      // Fan branches to each agent
      phaseMap[ph].forEach((_, fai) => {
        const fcol = fai % agCols
        const agCx = gsx + fcol * (AG_W + AG_COL_GAP) + AG_W / 2
        if (Math.abs(agCx - cx) < 2) {
          body += `<line x1="${cx}" y1="${dropMid}" x2="${agCx}" y2="${AGENT_START_Y}"
            stroke="${nodeStroke}" stroke-width="0.8" opacity="0.25"/>`
        } else {
          body += `<path d="M${cx},${dropMid} C${cx},${dropMid + 6} ${agCx},${AGENT_START_Y - 6} ${agCx},${AGENT_START_Y}"
            fill="none" stroke="${nodeStroke}" stroke-width="0.8" opacity="0.25"/>`
        }
      })
    }

    // ── Agent chips ───────────────────────────────────────────────────────
    const agents = phaseMap[ph]
    const count  = agents.length
    const { cols } = agentGridDims(ph)

    const totalW_chips = cols * AG_W + (cols - 1) * AG_COL_GAP
    const gridStartX   = cx - totalW_chips / 2

    agents.forEach((ag, ai) => {
      const col = ai % cols
      const row = Math.floor(ai / cols)
      const ax  = gridStartX + col * (AG_W + AG_COL_GAP)
      const ay  = AGENT_START_Y + row * (AG_H + AG_ROW_GAP)
      const ast = ag.status
      const persona = getPersona(ag.agent)

      // Chip fill/stroke by status
      const chipFill   = ast === 'running' ? 'rgba(245,158,11,0.16)'
                       : ast === 'done'    ? 'rgba(34,197,94,0.08)'
                       : ast === 'blocked' ? 'rgba(239,68,68,0.09)'
                       : 'rgba(39,53,73,0.9)'
      const chipStroke = ast === 'running' ? '#F59E0B'
                       : ast === 'done'    ? 'rgba(34,197,94,0.4)'
                       : ast === 'blocked' ? 'rgba(239,68,68,0.4)'
                       : 'rgba(51,65,85,0.8)'
      const chipStrokeW = ast === 'running' ? 2 : 1
      const chipAnim   = ast === 'running'
        ? `<animate attributeName="stroke-width" values="2;3;2" dur="1.6s" repeatCount="indefinite"/>
           <animate attributeName="opacity" values="1;0.82;1" dur="1.6s" repeatCount="indefinite"/>`
        : ''

      // Avatar — pixel art sprite if available, else colored circle with icon
      const avatarFill  = ps.avatar
      const avatarText  = ps.label
      const avatarURI   = getAvatarURI(ag.agent)
      const avatarSvg   = avatarURI
        ? `<circle cx="${ax + 22}" cy="${ay + 26}" r="14" fill="${avatarFill}" opacity="0.85"/>
           <image x="${ax + 8}" y="${ay + 12}" width="28" height="28"
             href="${avatarURI}"
             clip-path="url(#ac-${esc(ph)}-${ai})"
             style="image-rendering:pixelated"/>`
        : `<circle cx="${ax + 22}" cy="${ay + 26}" r="14" fill="${avatarFill}" opacity="0.9"/>
           <text x="${ax + 22}" y="${ay + 32}" text-anchor="middle"
             style="fill:${avatarText};font-family:system-ui,sans-serif;font-size:14px;pointer-events:none">
             ${esc(persona.icon)}</text>`

      // Status meta row
      const statusDot   = ast === 'done' ? '#22C55E' : ast === 'running' ? '#F59E0B' : ast === 'blocked' ? '#EF4444' : '#475569'
      const statusGlyph = ast === 'done' ? '✓' : ast === 'running' ? '▸' : ast === 'blocked' ? '✗' : ast === 'done_with_concerns' ? '!' : '○'
      const metaText    = ag.tokens > 0 ? fmtK(ag.tokens) + ' tok' : (ast === 'done' ? 'done' : ast)

      // One-liner summary (first sentence, max 36 chars)
      const rawSummary = (ag.summary || '').replace(/[#*`\n]/g, ' ').trim()
      const firstSentence = rawSummary.split(/[.!?]/)[0].trim()
      const snippetText = firstSentence.length > 36 ? firstSentence.slice(0, 34) + '…' : firstSentence

      // Tooltip (SVG <title> — shown on hover by browser)
      const tooltipFull = rawSummary.slice(0, 200) || `${persona.display} · ${ast}`

      // Left accent bar
      const accentBarColor = ast === 'running' ? '#F59E0B' : ast === 'done' ? '#22C55E' : ast === 'blocked' ? '#EF4444' : '#334155'

      body += `
        <g onclick="selectAgentByName('${esc(ag.agent)}')" style="cursor:pointer"
           role="button" tabindex="0" aria-label="${esc(persona.display)} (${esc(ag.agent)}): ${esc(ast)}">
          <title>${esc(tooltipFull)}</title>
          ${ast === 'running' ? `<rect x="${ax - 3}" y="${ay - 3}" width="${AG_W + 6}" height="${AG_H + 6}" rx="11" fill="rgba(245,158,11,0.06)" stroke="rgba(245,158,11,0.18)" stroke-width="1"/>` : ''}
          <rect x="${ax}" y="${ay}" width="${AG_W}" height="${AG_H}" rx="9"
            fill="${chipFill}" stroke="${chipStroke}" stroke-width="${chipStrokeW}"
            ${ast === 'running' ? 'filter="url(#glow-run)"' : ''}>${chipAnim}</rect>
          <rect x="${ax}" y="${ay + 10}" width="3" height="${AG_H - 20}" rx="1.5" fill="${accentBarColor}" opacity="0.8"/>
          ${avatarSvg}
          <text x="${ax + 42}" y="${ay + 21}" text-anchor="start"
            style="fill:#F1F5F9;font-family:system-ui,sans-serif;font-size:10.5px;font-weight:700;pointer-events:none">
            ${esc(persona.display)}</text>
          <text x="${ax + 42}" y="${ay + 33}" text-anchor="start"
            style="fill:#475569;font-family:system-ui,sans-serif;font-size:8.5px;pointer-events:none">
            ${esc(ag.agent)}</text>
          ${snippetText ? `<text x="${ax + 8}" y="${ay + 52}" text-anchor="start"
            style="fill:#94A3B8;font-family:system-ui,sans-serif;font-size:8px;pointer-events:none;font-style:italic">
            ${esc(snippetText)}</text>` : ''}
          <text x="${ax + 8}" y="${ay + 67}" text-anchor="start"
            style="fill:${statusDot};font-family:system-ui,sans-serif;font-size:8.5px;font-weight:600;pointer-events:none">
            ${statusGlyph} ${esc(metaText)}</text>
        </g>`
    })
  })

  svg.innerHTML = defs + body
}

function selectAgentByName(name) {
  const ag = state.agents.find(a => a.agent === name)
  if (ag) selectAgentData(ag)
}

function selectAgentData(ag) {
  if (!ag) return
  // Show/hide inspector empty state
  document.getElementById('agent-empty').style.display = 'none'
  const card = document.getElementById('active-card')
  card.style.display = 'flex'

  document.getElementById('ac-name').textContent = `${ag.agent} · ${ag.phase}`

  const pill = document.getElementById('ac-status-pill')
  pill.textContent = ag.status.replace(/_/g,' ')
  pill.className = `ac-status-pill ${ag.status}`

  document.getElementById('ac-summary').innerHTML = renderMarkdown(ag.summary || `${ag.status} — no output yet.`)
  document.getElementById('ac-conf').textContent = ag.conf ? `confidence: ${ag.conf}` : ''
  document.getElementById('ac-tokens').textContent = (ag.tokens > 0) ? fmtK(ag.tokens) + ' tokens' : ''

  const sources = ag.sources || []
  document.getElementById('ac-sources').innerHTML = sources.map(s =>
    `<a class="src-tag" href="${esc(s)}" target="_blank" rel="noopener" title="${esc(s)}">${
      esc(String(s).replace(/^https?:\/\//,'').slice(0,40))
    }</a>`
  ).join('')

  const delivEl = document.getElementById('ac-deliverables')
  const delivList = document.getElementById('ac-deliverables-list')
  if (ag.deliverables && ag.deliverables.length) {
    delivEl.style.display = 'flex'
    delivList.innerHTML = ag.deliverables.map(d =>
      `<div>├── ${esc(d)}</div>`
    ).join('')
  } else {
    delivEl.style.display = 'none'
  }

  // Switch to agent tab
  switchInspectorTab('agent')
}

// ── Orchestration card ───────────────────────────────────────────────────────
async function renderOrchestrationCard(runId) {
  const section = document.getElementById('orchestration-section')
  const body = document.getElementById('orchestration-body')
  const meta = document.getElementById('orchestration-meta')
  if (!section || !body) return

  try {
    const res = await fetch(`/api/runs/${runId}/files/approach.md`)
    if (!res.ok) { section.style.display = 'none'; return }
    const data = await res.json()
    const md = data.content || ''
    if (!md.trim()) { section.style.display = 'none'; return }

    // Extract context source and output destination badges
    const ctxMatch = md.match(/## Context Source\n- (.+)/)
    const outMatch = md.match(/## Output Configuration\n- Deliverable destination: (.+)/)
    const ctx = ctxMatch ? ctxMatch[1].trim() : null
    const out = outMatch ? outMatch[1].trim() : null

    if (meta) {
      meta.innerHTML = [
        ctx ? `<span class="orch-context-badge">📎 ${esc(ctx)}</span>` : '',
        out ? `<span class="orch-output-badge">→ ${esc(out)}</span>` : '',
      ].join('')
    }

    // Parse approach sections for card rendering
    const approachMatches = [...md.matchAll(/### Option (\d+)(\s*\(chosen\))?: (.+)\n\*\*Why[^*]*:\*\* (.+)\n\*\*Trade-off:\*\* (.+)/g)]
    let approachHtml = ''
    if (approachMatches.length > 0) {
      approachHtml = '<div class="orch-approaches">' +
        approachMatches.map(m => {
          const num = m[1], chosen = !!m[2], title = m[3], why = m[4], tradeoff = m[5]
          return `
            <div class="orch-approach-card${chosen ? ' chosen' : ''}">
              <div class="orch-approach-header">
                <span class="orch-approach-num">${esc(num)}</span>
                <span class="orch-approach-title">${esc(title)}</span>
                ${chosen ? '<span class="orch-chosen-badge">✓ chosen</span>' : ''}
              </div>
              <div class="orch-approach-why">${esc(why)}</div>
              <div class="orch-approach-tradeoff">Trade-off: ${esc(tradeoff)}</div>
            </div>`
        }).join('') + '</div>'
    } else {
      // Fallback: render full markdown
      approachHtml = `<div class="md-content">${renderMarkdown(md)}</div>`
    }

    // Parse clarifications Q&A
    const clarSection = md.match(/## Clarifications\n([\s\S]*?)(?=\n## )/)?.[1] || ''
    const clarHtml = clarSection ? `
      <div class="orch-qa-section">
        <div class="orch-qa-label">Clarifications</div>
        ${clarSection.trim().split('\n').map(line => {
          const m = line.match(/^- (.+?): (.+)$/)
          return m ? `<div class="orch-qa-pair"><span class="orch-q">${esc(m[1])}</span><span class="orch-a">${esc(m[2])}</span></div>` : ''
        }).join('')}
      </div>` : ''

    body.innerHTML = clarHtml + approachHtml
    section.style.display = 'block'
    body.classList.add('open')
    body.style.display = 'block'
    document.getElementById('orchestration-chevron').classList.add('open')
    document.querySelector('#orchestration-section .panel-header').classList.add('open')
  } catch (_) {
    section.style.display = 'none'
  }
}

// ── Phase outputs ─────────────────────────────────────────────────────────────
const PHASE_FILES = {
  planning:     [{ file: 'prd.md',             label: 'PRD' }],
  architecture: [{ file: 'adr.md',             label: 'ADR' },
                 { file: 'architecture.md',     label: 'Architecture' }],
  engineering:  [],
  qa:           [{ file: 'qa-report.md',        label: 'QA Report' },
                 { file: 'security-report.md',  label: 'Security Report' }],
  devops:       [{ file: 'review-report.md',    label: 'Code Review' }],
}

async function renderPhaseOutputs(runId) {
  const container = document.getElementById('phase-outputs')
  if (!container) return
  container.innerHTML = ''

  const donePhasesWithFiles = state.phases.filter(p =>
    p.status === 'done' && PHASE_FILES[p.phase_id] && PHASE_FILES[p.phase_id].length > 0
  )
  if (!donePhasesWithFiles.length) {
    await renderDocInspector(runId)
    return
  }

  for (const phase of donePhasesWithFiles) {
    const fileSpecs = PHASE_FILES[phase.phase_id]
    for (const spec of fileSpecs) {
      try {
        const res = await fetch(`/api/runs/${runId}/files/${encodeURIComponent(spec.file)}`)
        if (!res.ok) continue
        const data = await res.json()
        if (!data.content || !data.content.trim()) continue

        const panelId = `po-${phase.phase_id}-${spec.file.replace(/\./g,'-')}`
        const bodyId = `${panelId}-body`
        const chevronId = `${panelId}-chevron`

        const panel = document.createElement('div')
        panel.className = 'phase-output'
        panel.innerHTML = `
          <div class="phase-output-header open" onclick="togglePhaseOutput('${bodyId}','${chevronId}')">
            <span class="po-name">${esc(phase.phase_id)}</span>
            <span class="po-status">✓ done</span>
            <span class="po-timing">${esc(spec.label)}</span>
            <span class="po-badges">${badgesForPhase(phase.phase_id, data.content)}</span>
            <span class="po-chevron open" id="${chevronId}">▾</span>
          </div>
          <div class="phase-output-body open" id="${bodyId}">
            <div class="md-content">${renderMarkdown(data.content)}</div>
          </div>`
        container.appendChild(panel)
      } catch (_) {}
    }
  }

  // Also populate inspector docs tab
  await renderDocInspector(runId)
}

function togglePhaseOutput(bodyId, chevronId) {
  const body = document.getElementById(bodyId)
  if (!body) return
  const isOpen = body.classList.contains('open')
  body.classList.toggle('open', !isOpen)
  const chevron = document.getElementById(chevronId)
  if (chevron) chevron.classList.toggle('open', !isOpen)
}

function badgesForPhase(phaseId, content) {
  if (phaseId === 'qa') {
    const passMatch = content.match(/(\d+)\/(\d+) passing/)
    const bugMatch = content.match(/(\d+) bug/)
    let badges = ''
    if (passMatch) {
      const cls = passMatch[1] === passMatch[2] ? 'pass' : 'fail'
      badges += `<span class="po-badge ${cls}">${passMatch[1]}/${passMatch[2]} passing</span>`
    }
    if (bugMatch) {
      const cls = bugMatch[1] === '0' ? 'pass' : 'warn'
      badges += `<span class="po-badge ${cls}">${bugMatch[1]} bugs</span>`
    }
    return badges
  }
  if (phaseId === 'architecture') {
    if (/Accepted|APPROVED|Approved/i.test(content)) return '<span class="po-badge pass">APPROVED</span>'
  }
  if (phaseId === 'devops') {
    if (/APPROVED/i.test(content)) return '<span class="po-badge pass">APPROVED</span>'
    if (/CHANGES REQUESTED/i.test(content)) return '<span class="po-badge warn">CHANGES REQUESTED</span>'
  }
  return ''
}

// ── Inspector Docs tab ────────────────────────────────────────────────────────
async function renderDocInspector(runId) {
  const tabsEl = document.getElementById('doc-phase-tabs')
  const viewerEl = document.getElementById('doc-viewer')
  const emptyEl = document.getElementById('docs-empty')
  if (!tabsEl || !viewerEl) return

  // Collect available phase files
  const available = []
  const donePhasesWithFiles = state.phases.filter(p =>
    p.status === 'done' && PHASE_FILES[p.phase_id] && PHASE_FILES[p.phase_id].length > 0
  )

  for (const phase of donePhasesWithFiles) {
    for (const spec of PHASE_FILES[phase.phase_id]) {
      try {
        const res = await fetch(`/api/runs/${runId}/files/${encodeURIComponent(spec.file)}`)
        if (!res.ok) continue
        const data = await res.json()
        if (!data.content || !data.content.trim()) continue
        available.push({ phase: phase.phase_id, file: spec.file, label: spec.label, content: data.content })
      } catch (_) {}
    }
  }

  if (!available.length) {
    emptyEl.style.display = 'flex'
    tabsEl.style.display = 'none'
    viewerEl.innerHTML = ''
    return
  }

  emptyEl.style.display = 'none'
  tabsEl.style.display = 'flex'

  // Build sub-tabs
  tabsEl.innerHTML = available.map((item, i) =>
    `<button class="doc-ptab${i===0?' active':''}" onclick="selectDocTab(${i})" data-tab="${i}">${esc(item.label)}</button>`
  ).join('')

  // Store tabs data for click handler
  window._docTabs = available

  // Show first tab
  showDocTab(0)
}

function selectDocTab(i) {
  document.querySelectorAll('.doc-ptab').forEach((b,j) => b.classList.toggle('active', i===j))
  showDocTab(i)
}

function showDocTab(i) {
  const item = (window._docTabs || [])[i]
  const viewerEl = document.getElementById('doc-viewer')
  if (!item || !viewerEl) return

  // QA Report gets structured viewer
  if (item.phase === 'qa' && item.file === 'qa-report.md') {
    viewerEl.innerHTML = renderQAStructured(item.content)
    return
  }
  // Everything else: markdown
  viewerEl.innerHTML = `<div class="md-content">${renderMarkdown(item.content)}</div>`
}

// ── QA structured viewer ──────────────────────────────────────────────────────
function renderQAStructured(content) {
  let html = ''

  // 1. Parse layer table (look for markdown table with Layer|Method|Result columns)
  const tableMatch = content.match(/\|[\s]*Layer[\s]*\|[\s\S]*?(?=\n\n|\n#|$)/i)
  if (tableMatch) {
    const rows = tableMatch[0].trim().split('\n').filter(r => !/^\|[-:| ]+\|$/.test(r))
    if (rows.length > 1) {
      html += '<table class="doc-qa-table"><thead><tr>'
      const headers = rows[0].split('|').slice(1,-1).map(h => h.trim())
      headers.forEach(h => { html += `<th>${esc(h)}</th>` })
      html += '</tr></thead><tbody>'
      rows.slice(1).forEach(row => {
        const cells = row.split('|').slice(1,-1).map(c => c.trim())
        const firstCell = cells[0] || ''
        const isPassing = /✓|pass|✅/i.test(row)
        const isFail = /✗|fail|blocked|❌/i.test(row)
        const isSkip = /skip|n\/a|—/i.test(firstCell.toLowerCase())
        const rowClass = isPassing ? 'pass-row' : isFail ? 'fail-row' : ''
        html += `<tr class="${rowClass}">`
        cells.forEach(c => { html += `<td>${esc(c)}</td>` })
        html += '</tr>'
      })
      html += '</tbody></table>'
    }
  }

  // 2. Parse API test blocks — look for lines like:
  //    ✓ POST /auth/register → 201
  //    ✗ GET /users → expected 200, got 500
  const apiLines = content.match(/^[✓✗○]\s+(GET|POST|PUT|PATCH|DELETE)\s+\S+.*$/gm) || []
  if (apiLines.length) {
    html += '<div style="margin-bottom:14px"><div class="ad-section-label" style="margin-bottom:10px">API Tests</div>'
    apiLines.forEach((line, i) => {
      const pass = line.startsWith('✓')
      const m = line.match(/[✓✗○]\s+(GET|POST|PUT|PATCH|DELETE)\s+(\S+)\s*(→|expected|→)?\s*(\d+)?(.*)/)
      if (!m) return
      const method = m[1], path = m[2], statusCode = m[4] || ''
      const blockId = `atb-${i}`
      const statusCls = pass ? 's2xx' : 's4xx'
      html += `
        <div class="api-test-block">
          <div class="atb-header" onclick="toggleApiBlock('${blockId}')">
            <span class="atb-method ${method}">${esc(method)}</span>
            <span class="atb-path">${esc(path)}</span>
            ${statusCode ? `<span class="atb-status ${statusCls}">${esc(statusCode)}</span>` : ''}
            <span class="${pass ? 'pass-tag' : 'fail-tag'}" style="margin-left:8px;font-size:11px">${pass ? '✓ pass' : '✗ fail'}</span>
          </div>
          <div class="atb-body" id="${blockId}">
            <div class="atb-label">Raw</div>
            <div class="atb-req">${esc(line)}</div>
          </div>
        </div>`
    })
    html += '</div>'
  }

  // 3. Screenshots section
  html += `<div id="qa-screenshots"></div>`

  // 4. Fallback: if no structured content found, render as markdown
  if (!tableMatch && !apiLines.length) {
    html = `<div class="md-content">${renderMarkdown(content)}</div>`
  }

  return html
}

function toggleApiBlock(blockId) {
  const body = document.getElementById(blockId)
  if (!body) return
  const isOpen = body.classList.contains('open')
  body.classList.toggle('open', !isOpen)
}

// ── Utils ────────────────────────────────────────────────────────────────────
function esc(s) {
  return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;').replace(/'/g,'&#39;')
}
function fmtK(n) { return n >= 1000 ? (n / 1000).toFixed(1) + 'k' : String(n) }
function fmtTime(unix) {
  if (!unix) return '—'
  return new Date(unix * 1000).toLocaleTimeString()
}

document.addEventListener('DOMContentLoaded', init)
