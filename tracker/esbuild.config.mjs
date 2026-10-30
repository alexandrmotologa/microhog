import * as esbuild from 'esbuild';
import * as fs from 'fs';
import * as path from 'path';

const outDir = path.resolve('dist');
const serverDist = path.resolve('../internal/server/dist');

if (!fs.existsSync(outDir)) {
  fs.mkdirSync(outDir, { recursive: true });
}
if (!fs.existsSync(serverDist)) {
  fs.mkdirSync(serverDist, { recursive: true });
}

async function build() {
  console.log('Building tracker scripts...');

  // 1. Core Tracker
  const trackerResult = await esbuild.build({
    entryPoints: ['src/tracker.ts'],
    bundle: true,
    minify: true,
    target: ['es2020'],
    format: 'iife',
    outfile: path.join(outDir, 'microhog.js'),
    metafile: true,
  });

  const trackerSize = fs.statSync(path.join(outDir, 'microhog.js')).size;
  console.log(`✓ Core tracker built: ${(trackerSize / 1024).toFixed(2)} KB`);

  // Copy to server static dist
  fs.copyFileSync(path.join(outDir, 'microhog.js'), path.join(serverDist, 'microhog.js'));

  // 2. DOM Session Recorder
  await esbuild.build({
    entryPoints: ['src/recorder.ts'],
    bundle: true,
    minify: true,
    target: ['es2020'],
    format: 'iife',
    outfile: path.join(outDir, 'microhog-recorder.js'),
  });

  const recorderSize = fs.statSync(path.join(outDir, 'microhog-recorder.js')).size;
  console.log(`✓ Session recorder built: ${(recorderSize / 1024).toFixed(2)} KB`);

  fs.copyFileSync(path.join(outDir, 'microhog-recorder.js'), path.join(serverDist, 'microhog-recorder.js'));
  console.log('✓ Tracker assets synchronized with server dist');
}

build().catch((err) => {
  console.error(err);
  process.exit(1);
});
