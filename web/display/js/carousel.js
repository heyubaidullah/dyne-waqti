import { carouselSlide, idleFlyer, idleTimings } from './dom.js';

let slides = [];
let currentIndex = 0;
// 'flyer' | 'timings' — which full-screen page is currently showing while
// Idle. Countdown freezes whichever of the two was active without
// switching it (tick() is simply not called with isIdle=true then).
let phase = 'flyer';
let phaseStartedAtMs = null;
let timingsDurationSec = 15;
let lastIdsKey = '';

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function renderCurrentSlide() {
  if (slides.length === 0) {
    carouselSlide.innerHTML = '';
    return;
  }
  const slide = slides[currentIndex];
  if (slide.type === 'image') {
    // object-cover, not contain: a flyer must fill the entire screen
    // edge-to-edge (per spec), cropping as needed — the admin upload form
    // recommends 16:9 source images so cropping is rarely visible.
    carouselSlide.innerHTML = `<img src="${escapeHtml(slide.content_url_or_text)}" alt="${escapeHtml(slide.title)}" class="w-full h-full object-cover" />`;
  } else {
    const arabic = slide.arabic_text
      ? `<p class="font-arabic text-5xl mb-6" dir="rtl" lang="ar">${escapeHtml(slide.arabic_text)}</p>`
      : '';
    carouselSlide.innerHTML = `${arabic}<p class="text-4xl text-text-primary max-w-5xl leading-relaxed">${escapeHtml(slide.content_url_or_text)}</p>`;
  }
}

function showPhase(nextPhase) {
  phase = nextPhase;
  const showFlyer = phase === 'flyer' && slides.length > 0;
  idleFlyer.classList.toggle('hidden', !showFlyer);
  idleTimings.classList.toggle('hidden', showFlyer);
}

// Called on every data refresh (not per tick). Only replaces the slide
// list when the set of slide ids actually changed, so an unrelated
// field-only refresh doesn't interrupt whatever is currently showing —
// but does NOT reset the current phase/index, since a flyer naturally
// falling out of rotation (e.g. it expired) shouldn't visibly interrupt
// whatever is on screen at that exact moment either.
export function setSlides(newSlides) {
  const idsKey = JSON.stringify(newSlides.map((s) => s.id));
  if (idsKey === lastIdsKey) return;
  lastIdsKey = idsKey;
  slides = newSlides;
  if (currentIndex >= slides.length) currentIndex = 0;
}

export function setTimingsDuration(sec) {
  timingsDurationSec = sec > 0 ? sec : 15;
}

// Called once, by render.js, whenever the idle group (Idle/Countdown) is
// (re-)entered from any other state — always restarts the cycle at
// flyer-phase/slide-0 (or straight to timings if there are no active
// slides) rather than trying to resume a stale position after an
// interruption of unknown length.
export function restart(nowMs) {
  currentIndex = 0;
  phaseStartedAtMs = nowMs;
  if (slides.length > 0) {
    renderCurrentSlide();
    showPhase('flyer');
  } else {
    showPhase('timings');
  }
}

// Called every tick. Frozen (no-op) unless isIdle — Countdown explicitly
// pauses per spec; Silence/Blackout/Emergency hide the idle layer
// entirely via render.js, so ticking it would be wasted work anyway.
export function tick(nowMs, isIdle) {
  if (!isIdle) return;

  if (slides.length === 0) {
    if (phase !== 'timings') showPhase('timings');
    return;
  }
  if (phaseStartedAtMs === null) {
    phaseStartedAtMs = nowMs;
    return;
  }

  const durationMs =
    phase === 'flyer' ? (slides[currentIndex]?.display_duration_sec || 10) * 1000 : timingsDurationSec * 1000;

  if (nowMs - phaseStartedAtMs < durationMs) return;

  if (phase === 'flyer') {
    showPhase('timings');
  } else {
    currentIndex = (currentIndex + 1) % slides.length;
    renderCurrentSlide();
    showPhase('flyer');
  }
  phaseStartedAtMs = nowMs;
}
