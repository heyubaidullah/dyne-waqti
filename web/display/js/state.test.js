import { test } from 'node:test';
import assert from 'node:assert/strict';
import { computeState, jummahSlots, COUNTDOWN_LEAD_SEC } from './state.js';

function baseDisplayData(overrides = {}) {
  return {
    now: '2026-06-15T00:00:00+00:00',
    blackout: false,
    emergency: null,
    iqamah_times: {
      fajr: '06:00',
      dhuhr: '13:00',
      asr: '17:00',
      maghrib: '20:00',
      isha: '21:30',
    },
    ...overrides,
  };
}

// --- jummahSlots ---

test('jummahSlots: no slots hides both', () => {
  assert.deepEqual(jummahSlots([]), { jumuah: false, jumuah2: false });
  assert.deepEqual(jummahSlots(undefined), { jumuah: false, jumuah2: false });
});

test('jummahSlots: one slot shows only the first', () => {
  assert.deepEqual(jummahSlots([{ label: "Jumu'ah", iqamah: '13:00' }]), { jumuah: true, jumuah2: false });
});

test('jummahSlots: two slots shows both', () => {
  assert.deepEqual(
    jummahSlots([{ label: "Jumu'ah", iqamah: '13:00' }, { label: "Jumu'ah 2", iqamah: '14:15' }]),
    { jumuah: true, jumuah2: true },
  );
});

test('jummahSlots: going from two back to one cleanly drops the second', () => {
  const two = jummahSlots([{ label: 'a', iqamah: '13:00' }, { label: 'b', iqamah: '14:15' }]);
  assert.equal(two.jumuah2, true);
  const backToOne = jummahSlots([{ label: 'a', iqamah: '13:00' }]);
  assert.equal(backToOne.jumuah2, false);
  assert.equal(backToOne.jumuah, true);
});

// --- computeState: configurable silence-screen duration ---

test('silence screen is active just before the configured duration elapses (5min config)', () => {
  const data = baseDisplayData({ silence_duration_after_min: 5 });
  // 4 minutes after Fajr Iqamah (06:00) — inside a 5-minute window.
  const now = new Date('2026-06-15T06:04:00Z');
  const result = computeState(data, now);
  assert.equal(result.state, 'SILENCE');
  assert.equal(result.prayerName, 'fajr');
});

test('silence screen has cleared once the configured duration elapses (5min config)', () => {
  const data = baseDisplayData({ silence_duration_after_min: 5 });
  // 6 minutes after Fajr Iqamah — past the 5-minute window.
  const now = new Date('2026-06-15T06:06:00Z');
  const result = computeState(data, now);
  assert.notEqual(result.state, 'SILENCE');
});

test('a longer configured duration keeps the silence screen up longer (12min config)', () => {
  const data = baseDisplayData({ silence_duration_after_min: 12 });
  // 10 minutes after Fajr Iqamah — would have cleared under the 5min
  // config above, but must still be active under a 12min config.
  const now = new Date('2026-06-15T06:10:00Z');
  const result = computeState(data, now);
  assert.equal(result.state, 'SILENCE');
});

test('an unset/invalid silence duration falls back to the 7-minute default', () => {
  const data = baseDisplayData(); // no silence_duration_after_min at all
  const stillSilent = computeState(data, new Date('2026-06-15T06:06:00Z')); // 6min after
  assert.equal(stillSilent.state, 'SILENCE');
  const cleared = computeState(data, new Date('2026-06-15T06:08:00Z')); // 8min after
  assert.notEqual(cleared.state, 'SILENCE');
});

// --- computeState: sanity checks unrelated to this release, guarding the
// refactor from configurable-duration didn't change unrelated behavior ---

test('countdown state before Iqamah is unaffected by silence duration', () => {
  const data = baseDisplayData({ silence_duration_after_min: 5 });
  const now = new Date(new Date('2026-06-15T06:00:00Z').getTime() - (COUNTDOWN_LEAD_SEC - 60) * 1000);
  const result = computeState(data, now);
  assert.equal(result.state, 'COUNTDOWN');
  assert.equal(result.prayerName, 'fajr');
});

test('blackout and emergency still take priority regardless of silence duration', () => {
  const now = new Date('2026-06-15T06:04:00Z');
  assert.equal(computeState(baseDisplayData({ blackout: true, silence_duration_after_min: 5 }), now).state, 'BLACKOUT');
  assert.equal(
    computeState(baseDisplayData({ emergency: { deceased_name: 'x', prayer_time: 'y', location: 'z' }, silence_duration_after_min: 5 }), now).state,
    'EMERGENCY',
  );
});
