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

// ── Markdown renderer ────────────────────────────────────────────────────────
function renderMarkdown(raw) {
  if (!raw) return ''
  let s = raw
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')

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
  if (!ob || !rd) return
  ob.style.display = hasRuns ? 'none' : 'flex'
  rd.style.display = hasRuns ? 'block' : 'none'
}

function toggleSection(bodyId) {
  const body = document.getElementById(bodyId)
  if (!body) return
  body.classList.toggle('open')
  const sectionId = bodyId.replace('-body', '')
  const chevron = document.getElementById(sectionId + '-chevron')
  if (chevron) chevron.textContent = body.classList.contains('open') ? '▴' : '▾'
}

const NW = 150, NH = 44, GAP_X = 195, GAP_Y = 58, MX = 20, MY = 34

// ── Boot ────────────────────────────────────────────────────────────────────
async function init() {
  await loadWorkflows()
  await loadRuns()
  await loadStats()
  connectWS()
  renderTreeSimple()
  updateOnboardingVisibility()
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
    const dot = r.status === 'running' ? 'var(--amber)' :
                r.status === 'done'    ? 'var(--green)' :
                r.status === 'blocked' ? 'var(--red)'   : 'var(--muted)'
    const isActive = state.activeRun && r.id === state.activeRun.id
    const taskExcerpt = (r.task_text || '').slice(0, 40) + ((r.task_text || '').length > 40 ? '…' : '')
    return `
      <div class="run-item${isActive ? ' active' : ''}" data-id="${esc(r.id)}" onclick="loadRunDetail('${esc(r.id)}')">
        <div class="ri-header">
          <span class="status-badge" style="background:${dot}"></span>
          <span class="ri-name">${esc(r.workflow_name || r.id)}</span>
        </div>
        <div class="ri-meta">${esc(taskExcerpt || r.id)} · ${fmtTime(r.started_at)}</div>
      </div>`
  }).join('')
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

  if (!state.agents.length) {
    const w = svg.parentElement ? svg.parentElement.clientWidth || 700 : 700
    const isRunning = state.activeRun && (state.activeRun.status === 'running' || state.activeRun.status === 'pending')
    const noRuns = state.runs.length === 0

    if (isRunning) {
      svg.setAttribute('viewBox', `0 0 ${w} 100`)
      svg.setAttribute('height', '100')
      const isPending = state.activeRun.status === 'pending'
      svg.innerHTML = `
        <text x="${w/2}" y="42" text-anchor="middle"
          style="fill:${isPending ? 'var(--accent)' : 'var(--amber)'};font-family:var(--font);font-size:12px;font-weight:600">
          ${isPending ? 'Waiting for /team-dispatch in Claude Code…' : esc(state.activeRun.id) + ' — agents working…'}</text>
        <text x="${w/2}" y="62" text-anchor="middle"
          style="fill:var(--muted);font-family:var(--font);font-size:10px">
          ${isPending ? 'Run /team-dispatch --from-browser in your Claude session' : 'Tree updates automatically when first agent reports'}</text>`
      return
    }

    // noRuns case: onboarding HTML card is visible; tree-wrap is hidden — nothing to render
    if (noRuns) return

    svg.setAttribute('viewBox', `0 0 ${w} 80`)
    svg.setAttribute('height', '80')
    svg.innerHTML = `
      <text x="${w/2}" y="44" text-anchor="middle"
        style="fill:var(--muted);font-family:var(--font);font-size:11px">
        Select a run from the sidebar to view its agent tree.</text>`
    return
  }

  const phaseMap = {}
  state.agents.forEach(a => {
    if (!phaseMap[a.phase]) phaseMap[a.phase] = []
    phaseMap[a.phase].push(a)
  })
  const phases = Object.keys(phaseMap).sort((a, b) => phaseOrder(a) - phaseOrder(b))

  const phasePositions = {}
  phases.forEach((ph, pi) => {
    const x = MX + pi * GAP_X
    phasePositions[ph] = phaseMap[ph].map((ag, ai) => ({
      x, y: MY + ai * GAP_Y, cx: x + NW/2, cy: MY + ai * GAP_Y + NH/2, ag
    }))
  })

  const maxRows = Math.max(...phases.map(ph => phaseMap[ph].length))
  const totalH = MY + maxRows * GAP_Y + 30
  const totalW = MX + phases.length * GAP_X + NW + MX
  svg.setAttribute('viewBox', `0 0 ${totalW} ${totalH}`)
  svg.setAttribute('height', Math.max(totalH, 200))

  let html = '<defs>'
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi - 1]
    const prevNodes = phasePositions[prevPh]
    const currNodes = phasePositions[ph]
    const src = prevNodes[Math.floor(prevNodes.length / 2)]
    const dst = currNodes[Math.floor(currNodes.length / 2)]
    const x1 = src.x + NW, y1 = src.cy, x2 = dst.x, y2 = dst.cy
    const mx = (x1 + x2) / 2
    html += `<path id="sigpath-${pi}" style="display:none" d="M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}"/>`
  })
  html += '</defs>'

  // Edges
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi - 1]
    phasePositions[prevPh].forEach(src => {
      phasePositions[ph].forEach(dst => {
        const x1 = src.x + NW, y1 = src.cy, x2 = dst.x, y2 = dst.cy
        const mx = (x1 + x2) / 2
        html += `<path class="tree-edge" d="M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}"/>`
      })
    })
  })

  // Signal orbs
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const dur = (1.6 + pi * 0.25).toFixed(2)
    const delay = ((pi - 1) * 0.55).toFixed(2)
    html += `<circle class="signal-orb" r="3.5">
      <animateMotion dur="${dur}s" repeatCount="indefinite" begin="${delay}s">
        <mpath href="#sigpath-${pi}"/>
      </animateMotion>
    </circle>`
  })

  // Phase labels
  phases.forEach((ph, pi) => {
    const x = MX + pi * GAP_X + NW / 2
    html += `<text class="phase-label" x="${x}" y="${MY - 12}" text-anchor="middle">${esc(ph)}</text>`
  })

  // Nodes
  state.agents.forEach(ag => {
    const nodes = phasePositions[ag.phase]
    const ni = phaseMap[ag.phase].indexOf(ag)
    if (!nodes || ni < 0) return
    const { x, y } = nodes[ni]
    const st = ag.status
    const dot = st === 'done' ? '✓' : st === 'running' ? '▸' : st === 'blocked' ? '✗' : '○'
    const sub = ag.conf || (ag.tokens > 0 ? fmtK(ag.tokens) + ' tok' : st)
    html += `
      <rect class="node-rect ${esc(st)}" x="${x}" y="${y}" width="${NW}" height="${NH}" rx="5"
        onclick="selectAgentByName('${esc(ag.agent)}')" style="cursor:pointer" tabindex="0"
        role="button" aria-label="${esc(ag.agent)}: ${esc(st)}"/>
      <text class="node-name" x="${x + NW/2}" y="${y + 17}" text-anchor="middle" style="pointer-events:none">${esc(ag.agent)}</text>
      <text class="node-status-txt ${esc(st)}" x="${x + NW/2}" y="${y + 33}" text-anchor="middle" style="pointer-events:none">${esc(dot)} ${esc(sub)}</text>`
  })

  svg.innerHTML = html
}

function selectAgentByName(name) {
  const ag = state.agents.find(a => a.agent === name)
  if (ag) selectAgentData(ag)
}

function selectAgentData(ag) {
  if (!ag) return
  const card = document.getElementById('active-card')
  card.style.display = 'block'
  document.getElementById('ac-name').textContent = `${ag.agent} · ${ag.phase}`

  const pill = document.getElementById('ac-status-pill')
  pill.textContent = ag.status
  pill.className = `ac-status-pill ${ag.status}`

  document.getElementById('ac-summary').textContent = ag.summary || `${ag.status} — no output yet.`
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
    delivEl.style.display = 'block'
    delivList.innerHTML = ag.deliverables.map(d =>
      `<div class="ac-deliverables-list">├── ${esc(d)}</div>`
    ).join('')
  } else {
    delivEl.style.display = 'none'
  }
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
