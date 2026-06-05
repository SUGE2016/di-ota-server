import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const pkg = JSON.parse(readFileSync(join(root, 'package.json'), 'utf8'));
const infoPath = join(root, 'src', 'build-info.json');

let info = { version: pkg.version, build: 0 };
try {
  info = JSON.parse(readFileSync(infoPath, 'utf8'));
} catch {
  // first run
}

info.version = pkg.version;
info.build = Number(info.build || 0) + 1;

writeFileSync(infoPath, `${JSON.stringify(info, null, 2)}\n`, 'utf8');
console.log(`build-info: v${info.version}-build${info.build}`);
