#!/usr/bin/env node
import { readFileSync } from 'node:fs'
import { spawnSync } from 'node:child_process'

const registry = readFileSync('mcp-registry.yaml', 'utf8')
const packages = [...registry.matchAll(/args:\s*\["-y",\s*"([^"]+)"/g)].map(match => match[1])
const unique = [...new Set(packages)].sort()

let failed = false

for (const pkg of unique) {
  const result = spawnSync('npm', ['view', pkg, 'version'], { encoding: 'utf8' })
  if (result.status === 0) {
    process.stdout.write(`OK    ${pkg} ${result.stdout.trim()}\n`)
  } else {
    const firstLine = (result.stderr || result.stdout || 'unknown error').trim().split('\n')[0]
    process.stdout.write(`FAIL  ${pkg} ${firstLine}\n`)
    failed = true
  }
}

if (failed) {
  process.exit(1)
}
