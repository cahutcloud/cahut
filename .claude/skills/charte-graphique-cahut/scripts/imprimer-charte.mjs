// Imprime la version visuelle de la charte (index.html) en PDF A4 paysage, avec Chrome sans interface,
// à l'échelle 0,75 : chaque planche commence sur une nouvelle page.
//
// Usage : node .claude/skills/charte-graphique-cahut/scripts/imprimer-charte.mjs [sortie.pdf]
//   sortie.pdf  par défaut, docs/charte-graphique/charte-graphique-cahut.pdf
// Variables : CHROME=chemin pour un autre navigateur, ECHELLE pour une autre échelle. Node 22 ou plus.
import { spawn } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const racine = resolve(import.meta.dirname, '../../../..');
const source = join(racine, 'docs/charte-graphique/index.html');
const sortie = resolve(process.argv[2] ?? join(racine, 'docs/charte-graphique/charte-graphique-cahut.pdf'));
const echelle = Number(process.env.ECHELLE ?? 0.75);

const chrome = process.env.CHROME ?? [
  '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
  '/Applications/Chromium.app/Contents/MacOS/Chromium',
  '/usr/bin/google-chrome',
  '/usr/bin/chromium',
].find(existsSync);
if (!chrome) throw new Error('Chrome introuvable : définissez CHROME=chemin.');

// Profil jetable, pour ne pas toucher au Chrome ouvert de l'utilisateur.
const profil = mkdtempSync(join(tmpdir(), 'charte-'));
const port = 9300 + Math.floor(Math.random() * 600);
const navigateur = spawn(chrome, ['--headless=new', '--disable-gpu', '--allow-file-access-from-files',
  `--user-data-dir=${profil}`, `--remote-debugging-port=${port}`, 'about:blank'], { stdio: 'ignore' });

try {
  let cibles;
  for (let essai = 0; !cibles; essai++) {
    try {
      cibles = await (await fetch(`http://127.0.0.1:${port}/json/list`)).json();
    } catch (e) {
      if (essai > 50) throw e;
      await new Promise((r) => setTimeout(r, 100));
    }
  }
  const ws = new WebSocket(cibles.find((c) => c.type === 'page').webSocketDebuggerUrl);
  await new Promise((r) => (ws.onopen = r));
  let id = 0;
  const attentes = new Map();
  const evenements = [];
  ws.onmessage = (e) => {
    const m = JSON.parse(e.data);
    if (attentes.has(m.id)) attentes.get(m.id)(m);
    else if (m.method) evenements.forEach((f) => f(m));
  };
  const envoyer = (method, params = {}) => new Promise((r) => {
    attentes.set(++id, r);
    ws.send(JSON.stringify({ id, method, params }));
  });

  await envoyer('Page.enable');
  const charge = new Promise((r) => evenements.push((m) => m.method === 'Page.loadEventFired' && r()));
  await envoyer('Page.navigate', { url: pathToFileURL(source).href });
  await charge;
  await envoyer('Runtime.evaluate', { expression: 'document.fonts.ready.then(() => true)', awaitPromise: true });
  const { result, error } = await envoyer('Page.printToPDF', { printBackground: true, preferCSSPageSize: true, scale: echelle });
  if (error) throw new Error(error.message);
  writeFileSync(sortie, Buffer.from(result.data, 'base64'));
  ws.close();
  console.log(`${sortie} (échelle ${echelle})`);
} finally {
  const arret = new Promise((r) => navigateur.once('exit', r));
  navigateur.kill();
  await arret;
  rmSync(profil, { recursive: true, force: true, maxRetries: 5 });
}
