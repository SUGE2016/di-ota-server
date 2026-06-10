#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import { existsSync, readdirSync } from 'node:fs';
import { homedir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const mdToPdfBin = join(scriptDir, 'node_modules', '.bin', 'md-to-pdf');

function findChrome() {
  if (process.env.PUPPETEER_EXECUTABLE_PATH && existsSync(process.env.PUPPETEER_EXECUTABLE_PATH)) {
    return process.env.PUPPETEER_EXECUTABLE_PATH;
  }
  const root = join(homedir(), '.cache', 'puppeteer', 'chrome');
  if (!existsSync(root)) return '';
  for (const ver of readdirSync(root).sort().reverse()) {
    const bin = join(root, ver, 'chrome-linux64', 'chrome');
    if (existsSync(bin)) return bin;
  }
  return '';
}

const chrome = findChrome();
if (!chrome) {
  console.error('未找到 Puppeteer Chrome，请设置 PUPPETEER_EXECUTABLE_PATH');
  process.exit(1);
}
if (!existsSync(mdToPdfBin)) {
  console.error('请先执行: cd scripts && npm install');
  process.exit(1);
}

const pdfOptions = JSON.stringify({
  format: 'A4',
  margin: { top: '20mm', bottom: '20mm', left: '18mm', right: '18mm' },
  printBackground: true,
});

const launchOptions = JSON.stringify({
  args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage', '--disable-gpu'],
});

const env = { ...process.env, PUPPETEER_EXECUTABLE_PATH: chrome };
const files = process.argv.slice(2);
if (files.length === 0) {
  console.error('usage: node scripts/render-manual-pdf.mjs <file.md> [...]');
  process.exit(1);
}

for (const file of files) {
  const input = resolve(file);
  const output = input.replace(/\.md$/, '.pdf');
  process.stdout.write(`generating ${output} ... `);
  const started = Date.now();
  const result = spawnSync(
    mdToPdfBin,
    [input, '--pdf-options', pdfOptions, '--launch-options', launchOptions],
    { stdio: ['ignore', 'pipe', 'pipe'], env },
  );
  if (result.status !== 0) {
    process.stderr.write(result.stderr?.toString() || '');
    process.exit(result.status ?? 1);
  }
  console.log(`ok (${((Date.now() - started) / 1000).toFixed(1)}s)`);
}
