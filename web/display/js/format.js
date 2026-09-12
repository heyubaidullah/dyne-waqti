// Pure formatting helpers: no DOM access, fully unit-testable in isolation
// (render.js can't be imported directly under `node --test` since it touches
// `document` at module load — see js/dom.js).

// "MM:SS" under an hour (the Countdown overlay's own timer never exceeds
// ~12 minutes); rolls over to "H:MM:SS" beyond that, since the Idle view's
// "Next: X in ..." indicator can span many hours until the next prayer.
export function formatCountdown(seconds) {
  const s = Math.max(0, Math.round(seconds));
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  if (h > 0) {
    return `${h}:${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
  }
  return `${String(m).padStart(2, '0')}:${String(sec).padStart(2, '0')}`;
}
