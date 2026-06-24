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

const NW = 150, NH = 44, GAP_X = 195, GAP_Y = 58, MX = 20, MY = 34

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
  if (!svg) return

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

  const PH_W = 160, PH_H = 44, PH_GAP = 200  // phase node dimensions + center-to-center spacing
  const AG_W = 128, AG_H = 28, AG_GAP = 36    // agent chip dimensions + vertical gap
  const START_X = 24, PHASE_Y = 24             // starting positions
  const AGENT_START_Y = 84                     // agents start below phase node

  // Group agents by phase, in phase order
  const phaseMap = {}
  state.agents.forEach(a => {
    if (!phaseMap[a.phase]) phaseMap[a.phase] = []
    phaseMap[a.phase].push(a)
  })
  const phases = Object.keys(phaseMap).sort((a, b) => phaseOrder(a) - phaseOrder(b))

  // Phase status = worst-case agent status (or from state.phases)
  function phaseStatus(ph) {
    const statePhase = state.phases.find(p => p.phase_id === ph)
    if (statePhase) return statePhase.status || 'pending'
    const agents = phaseMap[ph] || []
    if (agents.some(a => a.status === 'blocked')) return 'blocked'
    if (agents.some(a => a.status === 'running')) return 'running'
    if (agents.every(a => a.status === 'done' || a.status === 'done_with_concerns')) return 'done'
    return 'pending'
  }

  // Compute positions
  const phaseX = {}  // phase -> center-x of phase node
  phases.forEach((ph, i) => { phaseX[ph] = START_X + PH_W/2 + i * PH_GAP })

  // Compute height — handle >4 agents in 2-row layout
  function agentRows(ph) {
    const count = phaseMap[ph].length
    return count > 4 ? 2 : 1
  }
  function agentsPerRow(ph) {
    const count = phaseMap[ph].length
    return count > 4 ? 2 : count
  }

  const maxRowsNeeded = Math.max(...phases.map(ph => {
    const count = phaseMap[ph].length
    return count > 4 ? Math.ceil(count / 2) : count
  }))

  const totalH = AGENT_START_Y + maxRowsNeeded * AG_GAP + 20
  const totalW = START_X + phases.length * PH_GAP + PH_W/2 + START_X

  svg.setAttribute('viewBox', `0 0 ${totalW} ${totalH}`)
  svg.setAttribute('height', Math.max(totalH, 200))

  let html = '<defs>'
  // Signal orb paths (between phase node centers)
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    const x1 = phaseX[prevPh] + PH_W/2, y1 = PHASE_Y + PH_H/2
    const x2 = phaseX[ph] - PH_W/2,    y2 = PHASE_Y + PH_H/2
    const mx = (x1+x2)/2
    html += `<path id="sigpath-${pi}" d="M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}" style="display:none"/>`
  })
  html += '</defs>'

  // Edges between phases
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    const x1 = phaseX[prevPh] + PH_W/2, y1 = PHASE_Y + PH_H/2
    const x2 = phaseX[ph] - PH_W/2,    y2 = PHASE_Y + PH_H/2
    const mx = (x1+x2)/2
    const dstStatus = phaseStatus(ph)
    const edgeCls = dstStatus === 'done' ? 'done' : dstStatus === 'running' ? 'running' : 'pending'
    html += `<path class="dag-edge tree-edge ${edgeCls}" d="M${x1},${y1} C${mx},${y1} ${mx},${y2} ${x2},${y2}"/>`
  })

  // Signal orbs (animate on running edges)
  phases.forEach((ph, pi) => {
    if (pi === 0) return
    const prevPh = phases[pi-1]
    const isRunning = phaseStatus(ph) === 'running' || phaseStatus(prevPh) === 'running'
    if (!isRunning) return
    const dur = (1.6 + pi * 0.3).toFixed(2)
    const delay = ((pi-1) * 0.5).toFixed(2)
    html += `<circle class="dag-signal-orb signal-orb" r="3.5">
      <animateMotion dur="${dur}s" repeatCount="indefinite" begin="${delay}s">
        <mpath href="#sigpath-${pi}"/>
      </animateMotion>
    </circle>`
  })

  // Phase nodes
  phases.forEach((ph, pi) => {
    const cx = phaseX[ph]
    const x = cx - PH_W/2, y = PHASE_Y
    const st = phaseStatus(ph)
    const dot = st === 'done' ? '✓' : st === 'running' ? '▸' : st === 'blocked' ? '✗' : ''
    html += `
      <rect class="dag-phase-node node-rect ${st}" x="${x}" y="${y}" width="${PH_W}" height="${PH_H}" rx="8"/>
      <text class="dag-phase-label phase-label" x="${cx}" y="${y + 27}" text-anchor="middle">
        ${dot ? esc(dot)+' ' : ''}${esc(ph)}
      </text>`

    // Agent chips below phase node
    const agents = phaseMap[ph]
    const agentCount = agents.length
    const useRows = agentCount > 4

    if (useRows) {
      // 2-column layout for >4 agents
      agents.forEach((ag, ai) => {
        const col = ai % 2
        const row = Math.floor(ai / 2)
        const totalRowW = 2 * AG_W + 8
        const agStartX = cx - totalRowW / 2
        const ax = agStartX + col * (AG_W + 8)
        const ay = AGENT_START_Y + row * AG_GAP
        const ast = ag.status
        const adot = ast === 'done' ? '✓' : ast === 'running' ? '▸' : ast === 'blocked' ? '✗' : ast === 'done_with_concerns' ? '⚠' : '○'
        const asub = ag.tokens > 0 ? fmtK(ag.tokens)+' tok' : adot
        html += `
          <rect class="dag-agent-chip node-rect ${ast}" x="${ax}" y="${ay}" width="${AG_W}" height="${AG_H}" rx="5"
            onclick="selectAgentByName('${esc(ag.agent)}')" style="cursor:pointer"
            role="button" tabindex="0" aria-label="${esc(ag.agent)}: ${esc(ast)}"/>
          <text class="dag-agent-label node-name" x="${ax+AG_W/2}" y="${ay+12}" text-anchor="middle" style="pointer-events:none">${esc(ag.agent)}</text>
          <text class="dag-agent-status node-status-txt ${ast}" x="${ax+AG_W/2}" y="${ay+24}" text-anchor="middle" style="pointer-events:none">${esc(asub)}</text>`
      })
    } else {
      // Single row — center agent chips below phase node
      const totalAgW = agentCount * AG_W + (agentCount - 1) * 8
      const agStartX = cx - totalAgW / 2
      agents.forEach((ag, ai) => {
        const ax = agStartX + ai * (AG_W + 8)
        const ay = AGENT_START_Y
        const ast = ag.status
        const adot = ast === 'done' ? '✓' : ast === 'running' ? '▸' : ast === 'blocked' ? '✗' : ast === 'done_with_concerns' ? '⚠' : '○'
        const asub = ag.tokens > 0 ? fmtK(ag.tokens)+' tok' : adot
        html += `
          <rect class="dag-agent-chip node-rect ${ast}" x="${ax}" y="${ay}" width="${AG_W}" height="${AG_H}" rx="5"
            onclick="selectAgentByName('${esc(ag.agent)}')" style="cursor:pointer"
            role="button" tabindex="0" aria-label="${esc(ag.agent)}: ${esc(ast)}"/>
          <text class="dag-agent-label node-name" x="${ax+AG_W/2}" y="${ay+12}" text-anchor="middle" style="pointer-events:none">${esc(ag.agent)}</text>
          <text class="dag-agent-status node-status-txt ${ast}" x="${ax+AG_W/2}" y="${ay+24}" text-anchor="middle" style="pointer-events:none">${esc(asub)}</text>`
      })
    }
  })

  svg.innerHTML = html
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
          <div class="phase-output-header" onclick="togglePhaseOutput('${bodyId}','${chevronId}')">
            <span class="po-name">${esc(phase.phase_id)}</span>
            <span class="po-status">✓ done</span>
            <span class="po-timing">${esc(spec.label)}</span>
            <span class="po-badges">${badgesForPhase(phase.phase_id, data.content)}</span>
            <span class="po-chevron" id="${chevronId}">▾</span>
          </div>
          <div class="phase-output-body" id="${bodyId}">
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
