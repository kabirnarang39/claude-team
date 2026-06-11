import Database from 'better-sqlite3'
import path from 'path'
import fs from 'fs'

function getDbPath() {
  const dir = process.env.CLAUDE_TEAM_DIR ?? path.join(process.cwd(), '.claude-team')
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true })
  return path.join(dir, 'messages.db')
}

export function createBus() {
  const db = new Database(getDbPath())
  db.exec(`
    CREATE TABLE IF NOT EXISTS messages (
      id        INTEGER PRIMARY KEY AUTOINCREMENT,
      from_role TEXT NOT NULL,
      to_role   TEXT NOT NULL,
      content   TEXT NOT NULL,
      timestamp INTEGER NOT NULL
    )
  `)
  return db
}

export function addMessage(db, fromRole, toRole, content) {
  db.prepare(
    'INSERT INTO messages (from_role, to_role, content, timestamp) VALUES (?, ?, ?, ?)'
  ).run(fromRole, toRole, content, Date.now())
}

export function getMessages(db, limit = 1000) {
  return db.prepare('SELECT * FROM messages ORDER BY timestamp ASC LIMIT ?').all(limit)
}

export function getRecentMessages(db, limit = 50) {
  return db.prepare(
    'SELECT * FROM messages ORDER BY timestamp DESC LIMIT ?'
  ).all(limit).reverse()
}
