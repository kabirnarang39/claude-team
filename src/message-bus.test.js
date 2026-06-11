import path from 'path'
import os from 'os'
import fs from 'fs'

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'claude-team-bus-'))
process.env.CLAUDE_TEAM_DIR = tmpDir

const { createBus, addMessage, getMessages, getRecentMessages } = await import('./message-bus.js')

let bus

beforeAll(() => { bus = createBus() })
afterAll(() => {
  bus.close()
  fs.rmSync(tmpDir, { recursive: true })
  delete process.env.CLAUDE_TEAM_DIR
})

test('addMessage stores a row', () => {
  addMessage(bus, 'manager', 'senior-architect', 'design auth schema')
  const msgs = getMessages(bus)
  expect(msgs.length).toBeGreaterThan(0)
  const last = msgs[msgs.length - 1]
  expect(last.from_role).toBe('manager')
  expect(last.to_role).toBe('senior-architect')
  expect(last.content).toBe('design auth schema')
})

test('getRecentMessages limits and orders results', () => {
  for (let i = 0; i < 5; i++) addMessage(bus, 'manager', 'engineer', `msg${i}`)
  const msgs = getRecentMessages(bus, 3)
  expect(msgs).toHaveLength(3)
  expect(msgs[0].timestamp).toBeLessThanOrEqual(msgs[1].timestamp)
})

test('all messages have integer timestamps', () => {
  getMessages(bus).forEach(m => expect(Number.isInteger(m.timestamp)).toBe(true))
})
