// Every element reference the app needs, queried exactly once at module
// load. Nothing else in the app calls querySelector on a per-tick basis —
// this is part of what keeps the 24/7 tick loop allocation-free.

export const logoEl = document.getElementById('logo');
export const waqtiLogoEl = document.getElementById('waqti-logo');
export const idleFlyer = document.getElementById('idle-flyer');
export const carouselSlide = document.getElementById('carousel-slide');
export const idleText = document.getElementById('idle-text');
export const textSlideContent = document.getElementById('text-slide-content');
export const idleTimings = document.getElementById('idle-timings');

// The prayer grid/clock/Hijri markup is duplicated in two places (the
// compact banner inside #idle-text, and the full-screen version inside
// #idle-timings) so both can be kept in sync from the same data with no
// per-view branching — these are NodeLists/arrays of 2 elements each,
// not single elements.
export const clockEls = document.querySelectorAll('.js-clock');
export const hijriDateEls = document.querySelectorAll('.js-hijri-date');
export const idleNextCountdownEls = document.querySelectorAll('.js-idle-next-countdown');

export const stateCountdown = document.getElementById('state-countdown');
export const countdownPrayerNameEl = document.getElementById('countdown-prayer-name');
export const countdownTimerEl = document.getElementById('countdown-timer');

export const stateSilence = document.getElementById('state-silence');
export const stateBlackout = document.getElementById('state-blackout');

export const stateEmergency = document.getElementById('state-emergency');
export const emergencyNameEl = document.getElementById('emergency-name');
export const emergencyTimeEl = document.getElementById('emergency-time');
export const emergencyLocationEl = document.getElementById('emergency-location');

// Same duplication as above: each prayer name maps to an array of 2
// elements (banner + full-screen), not a single element.
export const prayerCols = {};
for (const el of document.querySelectorAll('.prayer-col')) {
  const name = el.dataset.prayer;
  (prayerCols[name] ??= []).push(el);
}
