import { jest } from '@jest/globals'
import path from 'path'
import os from 'os'
import fs from 'fs'

const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'claude-team-reg-'))
process.env.CLAUDE_TEAM_DIR = tmpDir

const { getSessionId, getAllSessionIds, resetSessionId, ROLES_LIST } = await import('./registry.js')

afterAll(() => {
  fs.rmSync(tmpDir, { recursive: true })
  delete process.env.CLAUDE_TEAM_DIR
})

test('getSessionId returns same UUID for same role across calls', () => {
  const id1 = getSessionId('manager')
  const id2 = getSessionId('manager')
  expect(id1).toBe(id2)
})

test('getSessionId returns different UUID for different roles', () => {
  expect(getSessionId('manager')).not.toBe(getSessionId('senior-architect'))
})

test('getSessionId returns valid UUID v5 format', () => {
  const id = getSessionId('engineer')
  expect(id).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-5[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/)
})

test('getAllSessionIds returns entry for all 6 roles', () => {
  const all = getAllSessionIds()
  ROLES_LIST.forEach(role => expect(all[role]).toBeDefined())
})

test('registry persists to disk as JSON', () => {
  getSessionId('manager')
  const filePath = path.join(tmpDir, 'registry.json')
  expect(fs.existsSync(filePath)).toBe(true)
  const data = JSON.parse(fs.readFileSync(filePath, 'utf8'))
  expect(data.roles.manager).toBeDefined()
})

test('resetSessionId generates a new UUID', () => {
  const original = getSessionId('engineer')
  const reset = resetSessionId('engineer')
  expect(reset).not.toBe(original)
})
