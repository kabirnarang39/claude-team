// ── State ───────────────────────────────────────────────────────────────────
const state = {
  ws: null,
  runs: [],
  activeRun: null,
  phases: [],
  agents: [],
  workflows: [],
}

function updateOnboardingVisibility() {
  const hasRuns = state.runs && state.runs.length > 0
  const ob = document.getElementById('onboarding-card')
  const tw = document.getElementById('tree-wrap')
  if (!ob || !tw) return
  ob.style.display = hasRuns ? 'none' : 'flex'
  tw.style.display = hasRuns ? 'block' : 'none'
}

const NW = 150, NH = 44, GAP_X = 195, GAP_Y = 58, MX = 20, MY = 34

// ── Boot ────────────────────────────────────────────────────────────────────
async function init() {
  await loadWorkflows()
  await loadRuns()
  connectWS()
  bindEvents()
  renderTreeSimple()
  updateOnboardingVisibility()
  // Poll active run every 6s — agent results land in SQLite async
  setInterval(async () => {
    if (state.activeRun) await refreshActiveRun()
  }, 6000)
  // Token ticker — visually increment tokens only for actively running agents
  setInterval(() => {
    let changed = false
    state.agents.forEach(ag => {
      if (ag.status === 'running') {
        ag.tokens = (ag.tokens || 0) + Math.floor(Math.random() * 80 + 40)
        changed = true
      }
    })
    if (changed) renderTreeSimple()
  }, 1800)
}

async function refreshActiveRun() {
  try {
    const res = await fetch(`/api/runs/${state.activeRun.id}`)
    if (!res.ok) return
    const detail = await res.json()
    const prevCount = state.agents.length
    state.activeRun = detail
    state.phases = detail.phases || []
    const freshAgents = (detail.results || []).map(mapResult)
    // Preserve locally-inflated token counts for running agents not yet reported to DB
    state.agents = freshAgents.map(fresh => {
      if (fresh.tokens) return fresh
      const prev = state.agents.find(a => a.agent === fresh.agent && a.phase === fresh.phase)
      return (prev && prev.tokens) ? { ...fresh, tokens: prev.tokens } : fresh
    })
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
      if (evt.type === 'agent_result') onAgentResult(evt.payload)
    } catch (_) {}
  }
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
  }
}

// ── Data ─────────────────────────────────────────────────────────────────────
async function loadWorkflows() {
  try {
    const res = await fetch('/api/workflows')
    state.workflows = await res.json() || []
  } catch (_) { state.workflows = [] }
  renderWorkflowSelect()
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
    renderPhaseBar()
    renderTreeSimple()
    if (state.agents.length > 0) selectAgentData(state.agents[state.agents.length - 1])
    else document.getElementById('active-card').style.display = 'none'
  } catch (_) {}
}

// ── Dispatch ─────────────────────────────────────────────────────────────────
async function dispatch() {
  const text = document.getElementById('task-in').value.trim()
  const jiraUrl = document.getElementById('jira-in').value.trim()
  if (!text) { showCmd(null, 'Enter a task description first.', true); return }
  const workflow = document.getElementById('wf-select').value || 'feature-build'
  let runId
  try {
    const res = await fetch('/api/task', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text, jiraUrl, workflow }),
    })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    const data = await res.json()
    runId = data.run_id
  } catch (e) {
    document.getElementById('offline-banner').style.display = 'block'
    showCmd(null, `Server unreachable: ${e.message}`, true)
    return
  }
  // Refresh run list to show new run immediately
  try {
    const rres = await fetch('/api/runs')
    if (rres.ok) { state.runs = await rres.json() || []; renderRunHistory(); updateOnboardingVisibility() }
    if (runId) await loadRunDetail(runId)
  } catch (_) {}
  showCmd(`/team-dispatch --from-browser --workflow ${workflow}`)
}

function showCmd(cmd, errorMsg, isError) {
  const banner = document.getElementById('cmd-banner')
  const cmdText = document.getElementById('cmd-text')
  const label = banner.querySelector('.cmd-label')
  const copy = document.getElementById('cmd-copy')
  if (isError) {
    label.textContent = 'Error'
    label.style.color = 'var(--red)'
    banner.style.background = '#1a0808'
    banner.style.borderBottomColor = 'var(--red)'
    cmdText.textContent = errorMsg
    copy.style.display = 'none'
  } else {
    label.textContent = 'Run in Claude Code'
    label.style.color = 'var(--accent)'
    banner.style.background = '#0e0b1e'
    banner.style.borderBottomColor = 'var(--accent)'
    cmdText.textContent = cmd
    copy.style.display = ''
    copy.textContent = 'Copy'
    copy.className = 'cmd-copy'
  }
  banner.style.display = 'flex'
}

function dismissCmd() {
  document.getElementById('cmd-banner').style.display = 'none'
}

function copyCmd() {
  const txt = document.getElementById('cmd-text').textContent
  navigator.clipboard.writeText(txt).then(() => {
    const btn = document.getElementById('cmd-copy')
    btn.textContent = 'Copied'
    btn.className = 'cmd-copy copied'
    setTimeout(() => { btn.textContent = 'Copy'; btn.className = 'cmd-copy' }, 2000)
  })
}

// ── Renders ──────────────────────────────────────────────────────────────────
function renderWorkflowSelect() {
  const sel = document.getElementById('wf-select')
  if (!state.workflows.length) {
    sel.innerHTML = '<option value="feature-build">feature-build</option>'
    return
  }
  sel.innerHTML = state.workflows.map(w =>
    `<option value="${esc(w)}">${esc(w)}</option>`
  ).join('')
}

function renderRunHistory() {
  const el = document.getElementById('run-history')
  if (!state.runs.length) {
    el.innerHTML = '<div style="color:var(--muted);font-size:10px">No runs yet</div>'
    return
  }
  el.innerHTML = state.runs.map((r) => {
    const dot = r.status === 'running' ? 'var(--amber)' :
                r.status === 'done'    ? 'var(--green)' :
                r.status === 'blocked' ? 'var(--red)'   : 'var(--muted)'
    const isActive = state.activeRun && r.id === state.activeRun.id
    return `
      <div class="run-item${isActive ? ' active' : ''}" data-id="${esc(r.id)}" onclick="loadRunDetail('${esc(r.id)}')">
        <div class="ri-name">
          <span class="status-badge" style="background:${dot}"></span>${esc(r.workflow_name || r.id)}
        </div>
        <div class="ri-meta">${esc(r.status)} · ${fmtTime(r.started_at)}</div>
      </div>`
  }).join('')
}

function renderPhaseBar() {
  const bar = document.getElementById('phase-bar')
  if (!state.phases.length) { bar.innerHTML = ''; return }
  bar.innerHTML = state.phases.map((p, i) => `
    ${i > 0 ? '<span class="phase-arrow">→</span>' : ''}
    <span class="phase-pill ${esc(p.status || 'pending')}">${esc(p.phase_id)}</span>
  `).join('')
}

function phaseOrder(p) {
  const idx = state.phases.findIndex(ph => ph.phase_id === p)
  return idx >= 0 ? idx : 999
}

function renderTreeSimple() {
  const svg = document.getElementById('tree-svg')

  if (!state.agents.length) {
    const w = svg.parentElement ? svg.parentElement.clientWidth || 700 : 700
    const isRunning = state.activeRun && state.activeRun.status === 'running'
    const noRuns = state.runs.length === 0

    if (isRunning) {
      svg.setAttribute('viewBox', `0 0 ${w} 100`)
      svg.setAttribute('height', '100')
      svg.innerHTML = `
        <text x="${w/2}" y="42" text-anchor="middle"
          style="fill:var(--amber);font-family:var(--font);font-size:12px;font-weight:600">
          ${esc(state.activeRun.id)} — agents working…</text>
        <text x="${w/2}" y="62" text-anchor="middle"
          style="fill:var(--muted);font-family:var(--font);font-size:10px">
          Tree updates automatically when first agent reports</text>`
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
  document.getElementById('ac-summary').textContent = ag.summary || `${ag.status} — no output yet.`
  document.getElementById('ac-conf').textContent = ag.conf ? `confidence: ${ag.conf}` : ''
  document.getElementById('ac-tokens').textContent = ag.tokens ? fmtK(ag.tokens) + ' tokens' : ''
  const sources = ag.sources || []
  document.getElementById('ac-sources').innerHTML = sources.map(s =>
    `<a class="src-tag" href="${esc(s)}" target="_blank" rel="noopener noreferrer" title="${esc(s)}">${esc(String(s).replace(/^https?:\/\//, '').slice(0, 40))}</a>`
  ).join('')

  const outputEl = document.getElementById('ac-output')
  const outputContent = document.getElementById('ac-output-content')
  outputEl.style.display = 'none'
  if (state.activeRun) {
    fetch(`/api/runs/${state.activeRun.id}/files`)
      .then(r => r.ok ? r.json() : [])
      .then(files => {
        const agentSlug = ag.agent.toLowerCase().replace(/[-_]/g, '')
        const match = files.find(f => {
          const stem = f.toLowerCase().replace(/\.[^.]+$/, '').replace(/^report-/, '').replace(/[-_]/g, '')
          return stem === agentSlug
        })
        if (!match) return null
        return fetch(`/api/runs/${state.activeRun.id}/files/${encodeURIComponent(match)}`)
      })
      .then(r => r && r.ok ? r.json() : null)
      .then(data => {
        if (!data) return
        outputContent.textContent = data.content
        outputEl.style.display = 'block'
      })
      .catch(() => {})
  }
}

// ── Events ───────────────────────────────────────────────────────────────────
function bindEvents() {
  document.getElementById('dispatch-btn').addEventListener('click', dispatch)

  const dz = document.getElementById('drop-zone')
  dz.addEventListener('dragover', e => { e.preventDefault(); dz.classList.add('over') })
  dz.addEventListener('dragleave', () => dz.classList.remove('over'))
  dz.addEventListener('drop', async e => {
    e.preventDefault()
    dz.classList.remove('over')
    const file = e.dataTransfer.files[0]
    if (!file) return
    const data = await file.arrayBuffer()
    try {
      await fetch(`/api/files/upload?name=${encodeURIComponent(file.name)}`, { method: 'POST', body: data })
      dz.textContent = `✓ ${file.name} uploaded`
    } catch (_) { dz.textContent = 'Upload failed' }
  })
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
