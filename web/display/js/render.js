import * as dom from './dom.js';
import * as carousel from './carousel.js';
import { PRAYER_ORDER } from './state.js';

const PRAYER_LABELS = { fajr: 'Fajr', dhuhr: 'Dhuhr', asr: 'Asr', maghrib: 'Maghrib', isha: 'Isha' };

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
    if (result.state === 'COUNTDOWN') dom.stateCountdown.classList.remove('hidden');
    return;
  }

  // Leaving the idle group entirely — hide both full-screen idle pages
  // and the logo; the target state's own overlay takes over completely.
  dom.idleFlyer.classList.add('hidden');
  dom.idleTimings.classList.add('hidden');
  dom.logoEl.classList.add('hidden');

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

function formatCountdown(seconds) {
  const s = Math.max(0, Math.round(seconds));
  const m = Math.floor(s / 60);
  const sec = s % 60;
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
}

// Called every second — only touches textContent on already-cached
// elements, never queries the DOM or allocates persistent objects.
export function tickUpdate(result, now) {
  dom.clockEl.textContent = formatClock(now);

  if (result.state === 'IDLE') {
    dom.idleNextCountdownEl.textContent =
      result.nextPrayerName && result.secondsToIqamah != null
        ? `Next: ${PRAYER_LABELS[result.nextPrayerName]} in ${formatCountdown(result.secondsToIqamah)}`
        : '';
    for (const name of PRAYER_ORDER) {
      dom.prayerCols[name]?.classList.toggle('text-accent-gold', name === result.nextPrayerName);
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

  dom.hijriDateEl.textContent = `${data.hijri.day}/${data.hijri.month}/${data.hijri.year} AH`;
  for (const name of PRAYER_ORDER) {
    const col = dom.prayerCols[name];
    if (!col) continue;
    col.querySelector('.adhan-time').textContent = data.adhan_times[name];
    col.querySelector('.iqamah-time').textContent = data.iqamah_times[name];
  }
  if (dom.prayerCols.jumuah) {
    dom.prayerCols.jumuah.querySelector('.iqamah-time').textContent = data.iqamah_times.jumuah;
  }
}
