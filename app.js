"use strict";

const $ = (id) => document.getElementById(id);
const els = {
  frame: $("frame"), art: $("art"), loading: $("loading"), caption: $("caption"),
  seedForm: $("seed-form"), seed: $("seed"), newSeed: $("new-seed"), shuffle: $("shuffle"),
  styles: $("styles"), palettes: $("palettes"), density: $("density"), densityOut: $("density-out"),
  animate: $("animate"), dlSvg: $("dl-svg"), dlPng: $("dl-png"), copyLink: $("copy-link"),
  cli: $("cli"), copyCli: $("copy-cli"), status: $("status"),
};

const state = { seed: "", style: "", palette: "", density: 1, animate: false };
let current = null;          // last render result
let artURL = null;
const past = [];
let historyAt = -1;
let thumbTimer = 0;
let thumbToken = 0;

function pick(list) { return list[Math.floor(Math.random() * list.length)]; }

function readHash() {
  const h = new URLSearchParams(location.hash.slice(1));
  return {
    seed: h.get("seed") || "",
    style: h.get("style") || "",
    palette: h.get("palette") || "",
    density: parseFloat(h.get("density")) || 1,
  };
}

function writeHash() {
  const h = new URLSearchParams({ seed: state.seed, style: state.style, palette: state.palette });
  if (state.density !== 1) h.set("density", state.density.toFixed(1));
  window.history.replaceState(null, "", "#" + h);
}

function svgURL(svg) {
  return URL.createObjectURL(new Blob([svg], { type: "image/svg+xml" }));
}

function say(msg) {
  els.status.textContent = msg;
  clearTimeout(say.t);
  say.t = setTimeout(() => (els.status.textContent = ""), 2600);
}

function render({ remember = true } = {}) {
  const res = nacre.render({
    style: state.style, palette: state.palette, seed: state.seed,
    density: state.density, animate: state.animate, size: 1200,
  });
  if (res.error) {
    els.seed.setAttribute("aria-invalid", "true");
    say(res.error.replace(/^nacre: /, ""));
    return;
  }
  els.seed.removeAttribute("aria-invalid");
  current = res;
  state.seed = res.seed;
  state.style = res.style;
  state.palette = res.palette;

  if (artURL) URL.revokeObjectURL(artURL);
  artURL = svgURL(res.svg);
  els.art.src = artURL;
  els.art.alt = `${res.style} in the ${res.palette} palette, seed ${res.seed}`;
  els.frame.setAttribute("aria-busy", "false");
  els.caption.textContent = `${res.style}, ${res.palette}, ${res.shapes.toLocaleString()} shapes, drawn in ${res.ms} ms`;
  els.seed.value = res.seed;

  syncControls();
  writeHash();
  els.cli.textContent = cliCommand();
  if (remember) {
    const key = JSON.stringify([state.seed, state.style, state.palette, state.density]);
    if (past[historyAt] !== key) {
      past.splice(historyAt + 1);
      past.push(key);
      historyAt = past.length - 1;
    }
  }
  scheduleThumbs();
}

function cliCommand() {
  let cmd = `nacre render -style ${state.style} -seed ${state.seed} -palette ${state.palette}`;
  if (state.density !== 1) cmd += ` -density ${state.density.toFixed(1)}`;
  if (state.animate) cmd += " -animate";
  return cmd + ` -o ${state.style}.svg`;
}

function syncControls() {
  for (const input of els.styles.querySelectorAll("input")) input.checked = input.value === state.style;
  for (const input of els.palettes.querySelectorAll("input")) input.checked = input.value === state.palette;
  els.density.value = state.density;
  els.densityOut.textContent = state.density.toFixed(1);
}

// Draw the current seed in every style, a few frames apart so the main
// piece paints first and typing in the seed field stays responsive.
function scheduleThumbs() {
  clearTimeout(thumbTimer);
  const token = ++thumbToken;
  const opts = els.styles.querySelectorAll(".style-option");
  let i = 0;
  const step = () => {
    if (token !== thumbToken || i >= opts.length) return;
    const opt = opts[i++];
    const res = nacre.render({
      style: opt.dataset.style, palette: state.palette, seed: state.seed,
      density: Math.min(state.density, 1) * 0.55, size: 240,
    });
    if (!res.error) {
      const img = opt.querySelector("img");
      if (img.src) URL.revokeObjectURL(img.src);
      img.src = svgURL(res.svg);
    }
    thumbTimer = setTimeout(step, 16);
  };
  thumbTimer = setTimeout(step, 60);
}

function buildPickers() {
  for (const s of nacre.styles()) {
    const label = document.createElement("label");
    label.className = "style-option";
    label.dataset.style = s.name;
    label.title = s.info;
    label.innerHTML = `<input type="radio" name="style" value="${s.name}"><span class="thumb"><img alt=""></span><span class="name">${s.name}</span>`;
    label.querySelector("input").addEventListener("change", () => { state.style = s.name; render(); });
    els.styles.append(label);
  }
  for (const p of nacre.palettes()) {
    const label = document.createElement("label");
    label.className = "palette-option";
    const swatches = [p.background, ...p.colors].map((c) => `<i style="background:${c}"></i>`).join("");
    label.innerHTML = `<input type="radio" name="palette" value="${p.name}"><span class="strip" aria-hidden="true">${swatches}</span><span>${p.name}</span>`;
    label.querySelector("input").addEventListener("change", () => { state.palette = p.name; render(); });
    els.palettes.append(label);
  }
}

function download(name, blob) {
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = name;
  document.body.append(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(a.href), 1000);
}

function fileName(ext) { return `${state.style}-${state.seed}.${ext}`; }

function downloadSVG() {
  if (!current) return;
  download(fileName("svg"), new Blob([current.svg], { type: "image/svg+xml" }));
  say(`Saved ${fileName("svg")}`);
}

async function downloadPNG() {
  if (!current) return;
  els.dlPng.disabled = true;
  try {
    const still = nacre.render({ style: state.style, palette: state.palette, seed: state.seed, density: state.density, size: 1200 });
    const img = new Image();
    const url = svgURL(still.svg);
    await new Promise((ok, fail) => { img.onload = ok; img.onerror = fail; img.src = url; });
    const size = 2400;
    const canvas = document.createElement("canvas");
    canvas.width = canvas.height = size;
    canvas.getContext("2d").drawImage(img, 0, 0, size, size);
    URL.revokeObjectURL(url);
    const blob = await new Promise((ok) => canvas.toBlob(ok, "image/png"));
    download(fileName("png"), blob);
    say(`Saved ${fileName("png")} at ${size} × ${size}`);
  } catch {
    say("PNG export failed in this browser. Download the SVG instead.");
  } finally {
    els.dlPng.disabled = false;
  }
}

async function copy(text, what) {
  try {
    await navigator.clipboard.writeText(text);
    say(`Copied ${what}`);
  } catch {
    say(`Copy failed. Select the text and copy it by hand.`);
  }
}

function newSeed() { state.seed = nacre.seed(); render(); }

function shuffleAll() {
  state.seed = nacre.seed();
  state.style = pick(nacre.styles()).name;
  state.palette = pick(nacre.palettes()).name;
  render();
}

function goHistory(delta) {
  const next = historyAt + delta;
  if (next < 0 || next >= past.length) return;
  historyAt = next;
  [state.seed, state.style, state.palette, state.density] = JSON.parse(past[historyAt]);
  render({ remember: false });
}

function wire() {
  els.seedForm.addEventListener("submit", (e) => {
    e.preventDefault();
    const text = els.seed.value.trim();
    if (!text) { newSeed(); return; }
    const parsed = nacre.parseSeed(text);
    if (!parsed) { els.seed.setAttribute("aria-invalid", "true"); say("That seed can't be read. Try hex like 0x3f1a, or words."); return; }
    state.seed = text;
    render();
  });
  els.newSeed.addEventListener("click", newSeed);
  els.shuffle.addEventListener("click", shuffleAll);
  els.density.addEventListener("input", () => { els.densityOut.textContent = (+els.density.value).toFixed(1); });
  els.density.addEventListener("change", () => { state.density = +els.density.value; render(); });
  els.animate.addEventListener("change", () => { state.animate = els.animate.checked; render({ remember: false }); });
  els.dlSvg.addEventListener("click", downloadSVG);
  els.dlPng.addEventListener("click", downloadPNG);
  els.copyLink.addEventListener("click", () => copy(location.href, "link"));
  els.copyCli.addEventListener("click", () => copy(els.cli.textContent, "command"));
  window.addEventListener("hashchange", () => {
    const h = readHash();
    if (h.seed === state.seed && h.style === state.style && h.palette === state.palette && h.density === state.density) return;
    Object.assign(state, h);
    render();
  });
  document.addEventListener("keydown", (e) => {
    if (e.metaKey || e.ctrlKey || e.altKey) return;
    if (e.target.matches("input:not([type=range]):not([type=checkbox]):not([type=radio])")) return;
    if (e.key === "n" || e.key === "N") { e.preventDefault(); newSeed(); }
    else if (e.key === "ArrowLeft") { e.preventDefault(); goHistory(-1); }
    else if (e.key === "ArrowRight") { e.preventDefault(); goHistory(1); }
    else if (e.key === "s" || e.key === "S") { e.preventDefault(); downloadSVG(); }
  });
}

async function boot() {
  const go = new Go();
  const ready = new Promise((ok) => window.addEventListener("nacre-ready", ok, { once: true }));
  try {
    const src = fetch("nacre.wasm");
    const { instance } = WebAssembly.instantiateStreaming
      ? await WebAssembly.instantiateStreaming(src, go.importObject)
      : await WebAssembly.instantiate(await (await src).arrayBuffer(), go.importObject);
    go.run(instance);
    await ready;
  } catch (err) {
    els.loading.textContent = "The engine didn't load. Reload the page to try again.";
    console.error(err);
    return;
  }
  buildPickers();
  wire();
  const h = readHash();
  Object.assign(state, h);
  if (!state.seed) state.seed = nacre.seed();
  render();
}

boot();
