#!/usr/bin/env node
/**
 * claude-team coordinator MCP server
 * Provides ask/reply/inbox tools for inter-agent communication via SQLite.
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from '@modelcontextprotocol/sdk/types.js'
import Database from 'better-sqlite3'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const DB_PATH = process.env.ANTON_DB_PATH || path.join(process.cwd(), '.claude-team', 'state.db')

const db = new Database(DB_PATH)
db.pragma('journal_mode = WAL')

db.exec(`
  CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    read_at INTEGER,
    run_id TEXT,
    response TEXT
  );
  CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    workflow_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    started_at INTEGER NOT NULL,
    completed_at INTEGER
  );
  CREATE TABLE IF NOT EXISTS phases (
    run_id TEXT NOT NULL,
    phase_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    started_at INTEGER,
    completed_at INTEGER,
    UNIQUE(run_id, phase_id)
  );
  CREATE TABLE IF NOT EXISTS agent_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT NOT NULL,
    phase_id TEXT NOT NULL,
    agent TEXT NOT NULL,
    status TEXT NOT NULL,
    confidence TEXT DEFAULT 'medium',
    summary TEXT DEFAULT '',
    deliverables_json TEXT DEFAULT '[]',
    sources_json TEXT DEFAULT '[]',
    concerns_json TEXT DEFAULT '[]',
    questions_json TEXT DEFAULT '[]',
    tests_run TEXT DEFAULT '',
    tokens_used INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL
  );
`)

const server = new Server(
  { name: 'claude-team-coordinator', version: '1.4.5' },
  { capabilities: { tools: {} } }
)

server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools: [
    {
      name: 'ask',
      description: 'Send a question to another agent and wait for a response.',
      inputSchema: {
        type: 'object',
        properties: {
          from: { type: 'string', description: 'Your agent name (e.g. "engineer")' },
          target: { type: 'string', description: 'Target agent name (e.g. "manager", "senior-architect")' },
          question: { type: 'string', description: 'Your question' },
        },
        required: ['from', 'target', 'question'],
      },
    },
    {
      name: 'reply',
      description: 'Reply to a pending message in your inbox.',
      inputSchema: {
        type: 'object',
        properties: {
          message_id: { type: 'number', description: 'ID of the message you are replying to' },
          answer: { type: 'string', description: 'Your answer' },
        },
        required: ['message_id', 'answer'],
      },
    },
    {
      name: 'inbox',
      description: 'Check your unread messages from other agents.',
      inputSchema: {
        type: 'object',
        properties: {
          agent: { type: 'string', description: 'Your agent name' },
        },
        required: ['agent'],
      },
    },
    {
      name: 'coordinator_dispatch',
      description: 'Dispatch the next agents to run in the workflow. Call this when you have finished your current phase. The engine will spawn these agents and re-invoke you when they complete.',
      inputSchema: {
        type: 'object',
        properties: {
          agents: {
            type: 'array',
            items: { type: 'string' },
            description: 'Agent IDs to run next (must match IDs defined in the workflow YAML)',
          },
          reason: {
            type: 'string',
            description: 'Why you are dispatching these agents',
          },
        },
        required: ['agents'],
      },
    },
    {
      name: 'coordinator_finish',
      description: 'Signal that the workflow is complete. Call when all work is done and no more agents need to run.',
      inputSchema: {
        type: 'object',
        properties: {
          summary: {
            type: 'string',
            description: 'Final summary of what was accomplished',
          },
        },
        required: [],
      },
    },
    {
      name: 'report',
      description: 'Write structured task result to SQLite. Call before exiting — required for every agent.',
      inputSchema: {
        type: 'object',
        properties: {
          result: {
            type: 'string',
            description: 'JSON string: {agent, phase, status, confidence, summary, deliverables[], sources[], concerns[], questions[], tests_run, tokens_used}',
          },
        },
        required: ['result'],
      },
    },
  ],
}))

server.setRequestHandler(CallToolRequestSchema, async (req) => {
  const { name, arguments: args } = req.params

  if (name === 'ask') {
    const stmt = db.prepare(
      `INSERT INTO messages (from_agent, to_agent, content, created_at) VALUES (?, ?, ?, ?)`
    )
    const result = stmt.run(args.from, args.target, args.question, Date.now())
    return {
      content: [{
        type: 'text',
        text: `Message sent to ${args.target} (id: ${result.lastInsertRowid}). Check back with the inbox tool after a moment.`,
      }],
    }
  }

  if (name === 'inbox') {
    const rows = db.prepare(
      `SELECT id, from_agent, content, created_at FROM messages WHERE to_agent = ? AND read_at IS NULL ORDER BY created_at`
    ).all(args.agent)

    if (rows.length === 0) {
      return { content: [{ type: 'text', text: 'No new messages.' }] }
    }

    const msgs = rows.map(r =>
      `[id:${r.id}] From ${r.from_agent}: ${r.content}`
    ).join('\n')

    return { content: [{ type: 'text', text: msgs }] }
  }

  if (name === 'reply') {
    const msg = db.prepare(`SELECT * FROM messages WHERE id = ?`).get(args.message_id)
    if (!msg) return { content: [{ type: 'text', text: 'Message not found.' }] }

    db.prepare(`UPDATE messages SET read_at = ? WHERE id = ?`).run(Date.now(), args.message_id)

    db.prepare(
      `INSERT INTO messages (from_agent, to_agent, content, created_at) VALUES (?, ?, ?, ?)`
    ).run(msg.to_agent, msg.from_agent, `[Reply to id:${msg.id}] ${args.answer}`, Date.now())

    return { content: [{ type: 'text', text: `Reply sent to ${msg.from_agent}.` }] }
  }

  if (name === 'coordinator_dispatch') {
    return {
      content: [{
        type: 'text',
        text: `Dispatch acknowledged: [${args.agents.join(', ')}]. ${args.reason || ''} Write handoff file then exit.`,
      }],
    }
  }

  if (name === 'report') {
    let result
    try {
      result = JSON.parse(args.result)
    } catch (e) {
      console.error('[coordinator] ERROR: report called with invalid JSON:', e.message)
      return { content: [{ type: 'text', text: `Error: invalid JSON in result — ${e.message}` }] }
    }
    const runId = result.run_id || process.env.ANTON_RUN_ID || 'unknown'
    if (runId === 'unknown') {
      console.error('[coordinator] WARNING: report called with no run_id — results will be orphaned. Agent must include run_id in result JSON.')
    }
    const phaseId = result.phase_id || result.phase || 'unknown'
    const now = Math.floor(Date.now() / 1000)

    // Auto-create run row if first agent for this run
    db.prepare(`
      INSERT OR IGNORE INTO runs (id, workflow_name, status, started_at)
      VALUES (?, ?, 'running', ?)
    `).run(runId, process.env.ANTON_WORKFLOW || 'unknown', now)

    // Upsert phase row
    db.prepare(`
      INSERT INTO phases (run_id, phase_id, status, started_at)
      VALUES (?, ?, ?, ?)
      ON CONFLICT(run_id, phase_id) DO UPDATE SET status = excluded.status
    `).run(runId, phaseId, result.status === 'DONE' || result.status === 'DONE_WITH_CONCERNS' ? 'done' : 'running', now)

    db.prepare(`
      INSERT INTO agent_results
      (run_id, phase_id, agent, status, confidence, summary,
       deliverables_json, sources_json, concerns_json, questions_json,
       tests_run, tokens_used, created_at)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `).run(
      runId,
      phaseId,
      result.agent || 'unknown',
      result.status || 'DONE',
      result.confidence || 'medium',
      result.summary || '',
      JSON.stringify(result.deliverables || []),
      JSON.stringify(result.sources || []),
      JSON.stringify(result.concerns || []),
      JSON.stringify(result.questions || []),
      result.tests_run || '',
      result.tokens_used || 0,
      now
    )
    return { content: [{ type: 'text', text: 'reported' }] }
  }

  if (name === 'coordinator_finish') {
    const runId = process.env.ANTON_RUN_ID || 'unknown'
    if (runId === 'unknown') {
      console.error('[coordinator] WARNING: coordinator_finish called with no ANTON_RUN_ID — run status will not be updated.')
    }
    const now = Math.floor(Date.now() / 1000)
    db.prepare(`UPDATE runs SET status = 'done', completed_at = ? WHERE id = ?`).run(now, runId)
    return {
      content: [{
        type: 'text',
        text: `Workflow complete. ${args.summary || ''} You may exit now.`,
      }],
    }
  }

  throw new Error(`Unknown tool: ${name}`)
})

const transport = new StdioServerTransport()
await server.connect(transport)
