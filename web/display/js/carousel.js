import { carouselSlide, carouselViewport } from './dom.js';

let slides = [];
let currentIndex = 0;
let slideStartedAtMs = null;
let lastIdsKey = '';

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
}

function renderCurrent() {
  if (slides.length === 0) {
    carouselSlide.innerHTML = '';
    return;
  }
  const slide = slides[currentIndex];
  if (slide.type === 'image') {
    carouselSlide.innerHTML = `<img src="${escapeHtml(slide.content_url_or_text)}" alt="${escapeHtml(slide.title)}" class="max-h-full max-w-full object-contain" />`;
  } else {
    const arabic = slide.arabic_text
      ? `<p class="font-arabic text-4xl mb-4" dir="rtl" lang="ar">${escapeHtml(slide.arabic_text)}</p>`
      : '';
    carouselSlide.innerHTML = `${arabic}<p class="text-3xl text-text-primary max-w-4xl leading-relaxed">${escapeHtml(slide.content_url_or_text)}</p>`;
  }
}

// Called on every data refresh (not per tick). Only replaces the slide
// list — and resets to slide 0 — when the set of slide ids actually
// changed, so an unrelated field-only refresh doesn't interrupt whatever
// is currently showing.
export function setSlides(newSlides) {
  const idsKey = JSON.stringify(newSlides.map((s) => s.id));
  if (idsKey === lastIdsKey) return;
  lastIdsKey = idsKey;
  slides = newSlides;
  currentIndex = 0;
  slideStartedAtMs = null;
  renderCurrent();
}

// Called once, by render.js, whenever Idle is (re-)entered from any other
// state — always starts back at slide 0 rather than trying to resume a
// stale elapsed-time position after an interruption of unknown length.
export function restart(nowMs) {
  currentIndex = 0;
  slideStartedAtMs = nowMs;
  renderCurrent();
}

export function shrink(isShrunk) {
  carouselViewport.classList.toggle('h-[75%]', !isShrunk);
  carouselViewport.classList.toggle('h-[30%]', isShrunk);
}

// Called every tick. Frozen (no-op) unless isIdle — Countdown explicitly
// pauses rotation per spec; Silence/Blackout/Emergency hide the carousel
// entirely via render.js, so ticking it would be wasted work anyway.
export function tick(nowMs, isIdle) {
  if (!isIdle || slides.length === 0) return;
  if (slideStartedAtMs === null) {
    slideStartedAtMs = nowMs;
    renderCurrent();
    return;
  }
  const durationMs = (slides[currentIndex]?.display_duration_sec || 10) * 1000;
  if (nowMs - slideStartedAtMs >= durationMs) {
    currentIndex = (currentIndex + 1) % slides.length;
    slideStartedAtMs = nowMs;
    renderCurrent();
  }
}
