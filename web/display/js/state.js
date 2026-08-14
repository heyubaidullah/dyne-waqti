// Pure state-machine logic: no DOM access, no timers, fully unit-testable
// in isolation. computeState() is called once per tick with the last
// fetched display-data and the current best estimate of "now".

export const PRAYER_ORDER = ['fajr', 'dhuhr', 'asr', 'maghrib', 'isha'];

// Mid-range of the spec's stated windows (10-15min / 1-2min / 15-20min).
export const COUNTDOWN_LEAD_SEC = 12 * 60;
export const SILENCE_LEAD_SEC = 90;
export const SILENCE_DUR_SEC = 17 * 60;

// Extracts the server's local calendar date and UTC offset directly from
// the ISO timestamp string (e.g. "2026-08-14T02:30:16-04:00"), so the
// client never needs an IANA timezone database of its own — it just
// re-uses whatever offset the server (which does have one) already
// applied. This can be off by up to an hour on the one or two days per
// year the offset itself changes intraday (a DST transition landing
// between "now" and a later prayer the same day) — an accepted, narrow
// edge case for the display's live countdown; the backend's own prayer
// time calculation (internal/calc) is unaffected and remains correct.
function parseServerNow(nowIso) {
  const dateOnly = nowIso.slice(0, 10);
  const match = nowIso.match(/([+-]\d{2}:\d{2}|Z)$/);
  const offset = match ? (match[1] === 'Z' ? '+00:00' : match[1]) : '+00:00';
  return { dateOnly, offset };
}

function addDays(dateOnly, days) {
  const d = new Date(`${dateOnly}T00:00:00Z`);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

function prayerInstant(dateOnly, hhmm, offset) {
  return new Date(`${dateOnly}T${hhmm}:00${offset}`);
}

/**
 * @param {object} displayData - last fetched GET /api/v1/display-data body
 * @param {Date} estimatedNow - clock.estimatedNow()
 * @returns {{state: 'EMERGENCY'|'BLACKOUT'|'COUNTDOWN'|'SILENCE'|'IDLE', ...}}
 */
export function computeState(displayData, estimatedNow) {
  if (displayData.emergency) {
    return { state: 'EMERGENCY', emergency: displayData.emergency };
  }
  if (displayData.blackout) {
    return { state: 'BLACKOUT' };
  }

  const { dateOnly, offset } = parseServerNow(displayData.now);
  const tomorrow = addDays(dateOnly, 1);
  const nowMs = estimatedNow.getTime();

  const candidates = PRAYER_ORDER.map((name) => ({
    name,
    time: prayerInstant(dateOnly, displayData.iqamah_times[name], offset),
  }));
  // Wraparound: if every prayer today has already passed its silence
  // window, tomorrow's Fajr is the relevant "next" prayer.
  candidates.push({ name: 'fajr', time: prayerInstant(tomorrow, displayData.iqamah_times.fajr, offset) });

  let active = null;
  let nextFuture = null;
  for (const c of candidates) {
    const deltaSec = (c.time.getTime() - nowMs) / 1000; // positive = still upcoming
    if (deltaSec <= COUNTDOWN_LEAD_SEC && deltaSec >= -SILENCE_DUR_SEC) {
      if (!active || Math.abs(deltaSec) < Math.abs(active.deltaSec)) {
        active = { name: c.name, deltaSec };
      }
    }
    if (deltaSec > 0 && (!nextFuture || deltaSec < nextFuture.deltaSec)) {
      nextFuture = { name: c.name, deltaSec };
    }
  }

  if (active) {
    if (active.deltaSec > SILENCE_LEAD_SEC) {
      return { state: 'COUNTDOWN', prayerName: active.name, secondsToIqamah: active.deltaSec };
    }
    return { state: 'SILENCE', prayerName: active.name, secondsToIqamah: active.deltaSec };
  }

  return {
    state: 'IDLE',
    nextPrayerName: nextFuture ? nextFuture.name : null,
    secondsToIqamah: nextFuture ? nextFuture.deltaSec : null,
  };
}
