import { carouselSlide, idleFlyer, idleText, textSlideContent, idleTimings } from './dom.js';

let slides = [];
let currentIndex = 0;
// 'slide' | 'timings' — 'slide' shows whichever of #idle-flyer (image) or
// #idle-text (text_verse, with its own persistent banner) matches the
// current slide's type. Only image slides ever transition to 'timings':
// a text slide's banner already shows the prayer grid/clock/Hijri date
// simultaneously, so it never needs the full-screen interlude. Countdown
// freezes whichever view was active without switching it (tick() is
// simply not called with isIdle=true then).
let phase = 'slide';
let phaseStartedAtMs = null;
let timingsDurationSec = 15;
let lastIdsKey = '';

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function currentSlide() {
  return slides[currentIndex];
}

function renderCurrentSlide() {
  const slide = currentSlide();
  if (!slide) return;
  if (slide.type === 'image') {
    // object-cover, not contain: a flyer must fill the entire screen
    // edge-to-edge (per spec), cropping as needed — the admin upload form
    // recommends 16:9 source images so cropping is rarely visible.
    carouselSlide.innerHTML = `<img src="${escapeHtml(slide.content_url_or_text)}" alt="${escapeHtml(slide.title)}" class="w-full h-full object-cover" />`;
  } else {
    const arabic = slide.arabic_text
      ? `<p class="font-arabic text-5xl mb-6" dir="rtl" lang="ar">${escapeHtml(slide.arabic_text)}</p>`
      : '';
    textSlideContent.innerHTML = `${arabic}<p class="text-4xl text-text-primary max-w-5xl leading-relaxed">${escapeHtml(slide.content_url_or_text)}</p>`;
  }
}

// Shows whichever of #idle-flyer / #idle-text matches the current slide's
// type and hides the other two idle sub-views.
function showSlidePhase() {
  const slide = currentSlide();
  const isImage = Boolean(slide) && slide.type === 'image';
  idleFlyer.classList.toggle('hidden', !(slide && isImage));
  idleText.classList.toggle('hidden', !(slide && !isImage));
  idleTimings.classList.add('hidden');
}

function showTimingsPhase() {
  idleFlyer.classList.add('hidden');
  idleText.classList.add('hidden');
  idleTimings.classList.remove('hidden');
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
// slide-0 (or straight to timings if there are no active slides) rather
// than trying to resume a stale position after an interruption of
// unknown length.
export function restart(nowMs) {
  currentIndex = 0;
  phaseStartedAtMs = nowMs;
  if (slides.length > 0) {
    renderCurrentSlide();
    phase = 'slide';
    showSlidePhase();
  } else {
    phase = 'timings';
    showTimingsPhase();
  }
}

// Called every tick. Frozen (no-op) unless isIdle — Countdown explicitly
// pauses per spec; Silence/Blackout/Emergency hide the idle layer
// entirely via render.js, so ticking it would be wasted work anyway.
export function tick(nowMs, isIdle) {
  if (!isIdle) return;

  if (slides.length === 0) {
    if (phase !== 'timings') {
      phase = 'timings';
      showTimingsPhase();
    }
    return;
  }
  if (phaseStartedAtMs === null) {
    phaseStartedAtMs = nowMs;
    return;
  }

  const slide = currentSlide();
  const isImage = Boolean(slide) && slide.type === 'image';
  const durationMs = phase === 'slide' ? (slide?.display_duration_sec || 10) * 1000 : timingsDurationSec * 1000;

  if (nowMs - phaseStartedAtMs < durationMs) return;

  if (phase === 'slide' && isImage) {
    // Only images get the full-screen timings interlude — a text slide's
    // banner is already showing the prayer grid, so it advances straight
    // to the next slide below instead.
    phase = 'timings';
    showTimingsPhase();
  } else {
    // Skip the no-op re-render when a single text slide loops back to
    // itself — its content hasn't changed, so redrawing it would just be
    // a pointless flicker every cycle.
    if (slides.length > 1) {
      currentIndex = (currentIndex + 1) % slides.length;
      renderCurrentSlide();
    }
    phase = 'slide';
    showSlidePhase();
  }
  phaseStartedAtMs = nowMs;
}
