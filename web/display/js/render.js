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

function isBaseLayerState() {
  return currentStateName === 'IDLE' || currentStateName === 'COUNTDOWN';
}

function setBaseLayerVisible(visible) {
  dom.carouselViewport.classList.toggle('hidden', !visible);
  dom.infoBar.classList.toggle('hidden', !visible);
  dom.logoEl.classList.toggle('hidden', !visible || !hasLogo);
}

// Called every tick, but only acts when the state actually changed since
// the last tick — cheap string comparison, no-op on every other call.
export function applyState(result, nowMs) {
  if (result.state === currentStateName) return;
  const wasIdle = currentStateName === 'IDLE' || currentStateName === null;
  currentStateName = result.state;

  hideAllOverlays();

  switch (result.state) {
    case 'IDLE':
      setBaseLayerVisible(true);
      dom.infoBar.classList.remove('hidden');
      carousel.shrink(false);
      if (!wasIdle) carousel.restart(nowMs);
      break;
    case 'COUNTDOWN':
      setBaseLayerVisible(true);
      dom.infoBar.classList.add('hidden'); // grid hidden; carousel (shrunk) + countdown take over
      carousel.shrink(true);
      dom.stateCountdown.classList.remove('hidden');
      break;
    case 'SILENCE':
      setBaseLayerVisible(false);
      dom.stateSilence.classList.remove('hidden');
      break;
    case 'BLACKOUT':
      setBaseLayerVisible(false);
      dom.stateBlackout.classList.remove('hidden');
      break;
    case 'EMERGENCY':
      setBaseLayerVisible(false);
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
  dom.logoEl.classList.toggle('hidden', !hasLogo || !isBaseLayerState());

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
