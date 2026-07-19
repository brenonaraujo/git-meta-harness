#!/usr/bin/env node
// ============================================================================
// git-meta-harness — Playwright Visual Screenshot (v1.12.0)
// ============================================================================
// Helper para o frontend-engineer tirar screenshots locais ANTES do PR.
// Componente do sensor 12 (frontend-polish) + quality-assurance (Visual
// Report).
//
// Uso:
//   node harness/scripts/visual/playwright-screenshot.mjs
//   node harness/scripts/visual/playwright-screenshot.mjs --routes /,/auth/login,/dashboard
//   node harness/scripts/visual/playwright-screenshot.mjs --base-url http://localhost:3000
//   node harness/scripts/visual/playwright-screenshot.mjs --output qa/screenshots/
//
// Saída:
//   qa/screenshots/<route-slug>-<viewport>.png
//
// Dependências:
//   pnpm add -D playwright
//   pnpm exec playwright install chromium
// ============================================================================

import { chromium } from 'playwright';
import { mkdir } from 'node:fs/promises';
import { resolve } from 'node:path';

const args = process.argv.slice(2);
const getArg = (flag, defaultValue) => {
  const idx = args.indexOf(flag);
  return idx >= 0 && idx + 1 < args.length ? args[idx + 1] : defaultValue;
};

const baseUrl = getArg('--base-url', process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3000');
const routesArg = getArg('--routes', '/');
const outputDir = getArg('--output', 'qa/screenshots');
const viewportArg = getArg('--viewport', 'desktop');

const routes = routesArg.split(',').map(r => r.trim()).filter(Boolean);

const viewports = {
  mobile:  { width: 375,  height: 667 },
  tablet:  { width: 768,  height: 1024 },
  desktop: { width: 1440, height: 900 },
};

const slugify = (path) =>
  path.replace(/^\//, '').replace(/\//g, '_').replace(/[^a-zA-Z0-9_-]/g, '') || 'home';

async function main() {
  await mkdir(outputDir, { recursive: true });

  const viewport = viewports[viewportArg] ?? viewports.desktop;
  const browser = await chromium.launch();
  const context = await browser.newContext({ viewport });
  const page = await context.newPage();

  console.log(`==> Playwright visual screenshot (v1.12.0)`);
  console.log(`==> Base URL: ${baseUrl}`);
  console.log(`==> Viewport: ${viewportArg} (${viewport.width}x${viewport.height})`);
  console.log(`==> Routes:   ${routes.join(', ')}`);
  console.log(`==> Output:   ${outputDir}/`);
  console.log();

  for (const route of routes) {
    const url = `${baseUrl}${route}`;
    const slug = slugify(route);
    const filename = `${slug}-${viewportArg}.png`;
    const filepath = resolve(outputDir, filename);

    process.stdout.write(`  ${route.padEnd(30)} ... `);

    try {
      const response = await page.goto(url, { waitUntil: 'networkidle', timeout: 30000 });
      const status = response?.status() ?? 0;
      // Wait for fonts and CSS to settle
      await page.waitForLoadState('networkidle', { timeout: 5000 }).catch(() => {});
      await page.screenshot({ path: filepath, fullPage: true });
      console.log(`${status === 200 ? 'OK' : 'WARN'} (${status}) → ${filename}`);
    } catch (err) {
      console.log(`FAIL (${err.message})`);
    }
  }

  await browser.close();
  console.log();
  console.log(`Screenshots saved to ${outputDir}/`);
  console.log();
  console.log('Next steps:');
  console.log('  1. Open the screenshots and verify visual quality');
  console.log('  2. Compare with skill: harness/skills/visual-polish/SKILL.md');
  console.log('  3. Run sensor 12: ./harness/scripts/check-frontend-polish.sh');
  console.log('  4. If QA, generate visual report: qa/visual-report-<pr>.md');
}

main().catch(err => {
  console.error('FATAL:', err);
  process.exit(1);
});
