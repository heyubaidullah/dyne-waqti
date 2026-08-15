import * as dom from './dom.js';
import * as carousel from './carousel.js';
import { PRAYER_ORDER } from './state.js';

const PRAYER_LABELS = { fajr: 'Fajr', dhuhr: 'Dhuhr', asr: 'Asr', maghrib: 'Maghrib', isha: 'Isha' };

// 1-indexed by Hijri month number (index 0 unused) so HIJRI_MONTHS[data.hijri.month] reads directly.
const HIJRI_MONTHS = [
  '',
  'Muharram',
  'Safar',
  "Rabi' al-Awwal",
  "Rabi' al-Thani",
  'Jumada al-Awwal',
  'Jumada al-Thani',
  'Rajab',
  "Sha'ban",
  'Ramadan',
  'Shawwal',
  "Dhu al-Qi'dah",
  'Dhu al-Hijjah',
];

let currentStateName = null;
// The mosque's configured IANA timezone (from display-data), NOT the kiosk
// machine's own system timezone — a misconfigured or unset OS timezone on
// bare kiosk hardware must never affect what clock time is shown.
let currentTimezone = 'UTC';
// Whether a logo has been uploaded — the <img> only ever shows alongside
// the rest of the base layer (Idle/Countdown), never during
// Silence/Blackout/Emergency.
let hasLogo = false;

function hideAllOverlays() {
  dom.stateCountdown.classList.add('hidden');
  dom.stateSilence.classList.add('hidden');
  dom.stateBlackout.classList.add('hidden');
  dom.stateEmergency.classList.add('hidden');
}

// Idle and Countdown together form the "idle group": Countdown is just a
// semi-transparent overlay on top of whichever full-screen idle phase
// (flyer or timings) carousel.js already has showing, frozen in place —
// it never hides or resets that layer. Only entering/leaving the group
// as a whole touches the idle layer or the logo.
function isIdleGroup(stateName) {
  return stateName === 'IDLE' || stateName === 'COUNTDOWN';
}

// Called every tick, but only acts when the state actually changed since
// the last tick — cheap string comparison, no-op on every other call.
export function applyState(result, nowMs) {
  if (result.state === currentStateName) return;
  const wasInIdleGroup = isIdleGroup(currentStateName);
  const enteringIdleGroup = isIdleGroup(result.state);
  currentStateName = result.state;

  hideAllOverlays();

  if (enteringIdleGroup) {
    if (!wasInIdleGroup) carousel.restart(nowMs);
    dom.logoEl.classList.toggle('hidden', !hasLogo);
    dom.waqtiLogoEl.classList.remove('hidden');
    if (result.state === 'COUNTDOWN') dom.stateCountdown.classList.remove('hidden');
    return;
  }

  // Leaving the idle group entirely — hide all three idle sub-views and
  // both logos; the target state's own overlay takes over completely.
  dom.idleFlyer.classList.add('hidden');
  dom.idleText.classList.add('hidden');
  dom.idleTimings.classList.add('hidden');
  dom.logoEl.classList.add('hidden');
  dom.waqtiLogoEl.classList.add('hidden');

  switch (result.state) {
    case 'SILENCE':
      dom.stateSilence.classList.remove('hidden');
      break;
    case 'BLACKOUT':
      dom.stateBlackout.classList.remove('hidden');
      break;
    case 'EMERGENCY':
      dom.stateEmergency.classList.remove('hidden');
      dom.emergencyNameEl.textContent = result.emergency.deceased_name;
      dom.emergencyTimeEl.textContent = result.emergency.prayer_time;
      dom.emergencyLocationEl.textContent = result.emergency.location;
      break;
  }
}

function formatClock(d) {
  return d.toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: true,
    timeZone: currentTimezone,
  });
}

// "13:12" -> "1:12 PM" — the live clock was already 12h; the prayer-grid
// Adhan/Iqamah labels come straight from the API as 24h HH:MM and need
// the same treatment.
function formatTime12h(hhmm) {
  const match = /^(\d{1,2}):(\d{2})$/.exec(hhmm || '');
  if (!match) return hhmm || '';
  let h = Number(match[1]);
  const period = h >= 12 ? 'PM' : 'AM';
  h = h % 12;
  if (h === 0) h = 12;
  return `${h}:${match[2]} ${period}`;
}

function formatCountdown(seconds) {
  const s = Math.max(0, Math.round(seconds));
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
}

// Called every second — only touches textContent on already-cached
// elements, never queries the DOM or allocates persistent objects.
export function tickUpdate(result, now) {
  const clockText = formatClock(now);
  for (const el of dom.clockEls) el.textContent = clockText;

  if (result.state === 'IDLE') {
    const nextCountdownText =
      result.nextPrayerName && result.secondsToIqamah != null
        ? `Next: ${PRAYER_LABELS[result.nextPrayerName]} in ${formatCountdown(result.secondsToIqamah)}`
        : '';
    for (const el of dom.idleNextCountdownEls) el.textContent = nextCountdownText;
    for (const name of PRAYER_ORDER) {
      const isNext = name === result.nextPrayerName;
      for (const col of dom.prayerCols[name] || []) {
        col.classList.toggle('text-accent-gold', isNext);
        col.classList.toggle('bg-surface-slate', isNext);
        col.classList.toggle('ring-2', isNext);
        col.classList.toggle('ring-accent-gold', isNext);
      }
    }
  }

  if (result.state === 'COUNTDOWN') {
    dom.countdownPrayerNameEl.textContent = `${PRAYER_LABELS[result.prayerName]} Iqamah`;
    dom.countdownTimerEl.textContent = `IQAMAH IN ${formatCountdown(result.secondsToIqamah)}`;
  }
}

// Called only on data refresh (SSE event or the 60s poll) — fields that
// never change second-to-second have no business being touched by the
// 1s tick loop.
export function updateStaticFields(data) {
  currentTimezone = data.timezone;

  hasLogo = Boolean(data.logo_url);
  if (hasLogo) dom.logoEl.src = data.logo_url;
  // Re-sync immediately rather than waiting for the next state
  // transition — a logo can be uploaded/removed while already Idle.
  dom.logoEl.classList.toggle('hidden', !hasLogo || !isIdleGroup(currentStateName));

  carousel.setTimingsDuration(data.timings_duration_sec);

  const hijriMonthName = HIJRI_MONTHS[data.hijri.month] || data.hijri.month;
  const hijriText = `${data.hijri.day} ${hijriMonthName} ${data.hijri.year} AH`;
  for (const el of dom.hijriDateEls) el.textContent = hijriText;
  for (const name of PRAYER_ORDER) {
    for (const col of dom.prayerCols[name] || []) {
      col.querySelector('.adhan-time').textContent = formatTime12h(data.adhan_times[name]);
      col.querySelector('.iqamah-time').textContent = formatTime12h(data.iqamah_times[name]);
    }
  }
  for (const col of dom.prayerCols.jumuah || []) {
    col.querySelector('.iqamah-time').textContent = formatTime12h(data.iqamah_times.jumuah);
  }
}
