// Anchors to the server's authoritative "now" (already in the configured
// IANA timezone) on each data refresh, then advances the estimate using
// the client's own monotonic clock. Never trusts the client's wall clock
// or timezone directly — a kiosk with a wrong system clock or timezone
// must not throw off the prayer-time state machine.
let anchorServerMs = null;
let anchorPerfMs = null;

export function setAnchor(nowIso) {
  anchorServerMs = Date.parse(nowIso);
  anchorPerfMs = performance.now();
}

export function estimatedNow() {
  if (anchorServerMs === null) return new Date();
  return new Date(anchorServerMs + (performance.now() - anchorPerfMs));
}
