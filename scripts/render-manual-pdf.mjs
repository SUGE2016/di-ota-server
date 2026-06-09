#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import { resolve } from 'node:path';

const pdfOptions = JSON.stringify({
  format: 'A4',
  margin: { top: '20mm', bottom: '20mm', left: '18mm', right: '18mm' },
  printBackground: true,
});

const launchOptions = JSON.stringify({
  args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage'],
});

const files = process.argv.slice(2);
if (files.length === 0) {
  console.error('usage: node scripts/render-manual-pdf.mjs <file.md> [...]');
  process.exit(1);
}

for (const file of files) {
  const input = resolve(file);
  process.stdout.write(`generating ${input.replace(/\.md$/, '.pdf')} ... `);
  const result = spawnSync(
    'npx',
    ['--yes', 'md-to-pdf', input, '--pdf-options', pdfOptions, '--launch-options', launchOptions],
    { stdio: 'inherit', shell: false },
  );
  if (result.status !== 0) {
    process.exit(result.status ?? 1);
  }
  console.log('ok');
}
