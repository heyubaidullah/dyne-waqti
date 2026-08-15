// Every element reference the app needs, queried exactly once at module
// load. Nothing else in the app calls querySelector on a per-tick basis —
// this is part of what keeps the 24/7 tick loop allocation-free.

export const logoEl = document.getElementById('logo');
export const idleFlyer = document.getElementById('idle-flyer');
export const carouselSlide = document.getElementById('carousel-slide');
export const idleTimings = document.getElementById('idle-timings');
export const clockEl = document.getElementById('clock');
export const hijriDateEl = document.getElementById('hijri-date');
export const idleNextCountdownEl = document.getElementById('idle-next-countdown');

export const stateCountdown = document.getElementById('state-countdown');
export const countdownPrayerNameEl = document.getElementById('countdown-prayer-name');
export const countdownTimerEl = document.getElementById('countdown-timer');

export const stateSilence = document.getElementById('state-silence');
export const stateBlackout = document.getElementById('state-blackout');

export const stateEmergency = document.getElementById('state-emergency');
export const emergencyNameEl = document.getElementById('emergency-name');
export const emergencyTimeEl = document.getElementById('emergency-time');
export const emergencyLocationEl = document.getElementById('emergency-location');

export const prayerCols = Object.fromEntries(
  Array.from(document.querySelectorAll('.prayer-col')).map((el) => [el.dataset.prayer, el]),
);
