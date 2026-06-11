import fs from 'fs'
import path from 'path'
import { v5 as uuidv5 } from 'uuid'

const NAMESPACE = '6ba7b810-9dad-11d1-80b4-00c04fd430c8'

export const ROLES_LIST = [
  'manager',
  'senior-architect',
  'senior-engineer',
  'engineer',
  'peer-programmer',
  'qa'
]

function getTeamDir() {
  return process.env.CLAUDE_TEAM_DIR ?? path.join(process.cwd(), '.claude-team')
}

function getRegistryPath() {
  return path.join(getTeamDir(), 'registry.json')
}

function loadRegistry() {
  const filePath = getRegistryPath()
  if (fs.existsSync(filePath)) {
    return JSON.parse(fs.readFileSync(filePath, 'utf8'))
  }
  return { namespace: NAMESPACE, roles: {}, created: new Date().toISOString() }
}

function saveRegistry(data) {
  const dir = getTeamDir()
  if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true })
  fs.writeFileSync(getRegistryPath(), JSON.stringify(data, null, 2))
}

export function getSessionId(role) {
  const reg = loadRegistry()
  if (!reg.roles[role]) {
    reg.roles[role] = uuidv5(role, NAMESPACE)
    saveRegistry(reg)
  }
  return reg.roles[role]
}

export function getAllSessionIds() {
  return Object.fromEntries(ROLES_LIST.map(role => [role, getSessionId(role)]))
}

export function resetSessionId(role) {
  const reg = loadRegistry()
  const salt = `${role}-reset-${Object.keys(reg.roles).length}`
  reg.roles[role] = uuidv5(salt, NAMESPACE)
  saveRegistry(reg)
  return reg.roles[role]
}
