import { refreshDisplayData, connectSSE, getCachedDisplayData } from './api.js';
import { estimatedNow } from './clock.js';
import { computeState } from './state.js';
import { applyState, tickUpdate } from './render.js';
import { tick as carouselTick } from './carousel.js';

// Safety-net refresh interval: catches midnight day-rollover (new Iqamah
// times) and self-heals from any missed SSE message. Independent of the
// 1s tick loop, which only recomputes state from already-cached data.
const POLL_INTERVAL_MS = 60000;
const TICK_INTERVAL_MS = 1000;

async function safeRefresh() {
  try {
    await refreshDisplayData();
  } catch (err) {
    // Network hiccup or server restart — the next poll tick or SSE
    // reconnect retries automatically; nothing to clean up here, and
    // nothing accumulates on repeated failure.
    console.error('refreshDisplayData failed', err);
  }
}

async function init() {
  await safeRefresh();

  // The one and only tick loop for the page's entire lifetime — created
  // once, never cleared, never duplicated.
  setInterval(() => {
    const data = getCachedDisplayData();
    if (!data) return;
    const now = estimatedNow();
    const nowMs = now.getTime();
    const result = computeState(data, now);
    applyState(result, nowMs);
    tickUpdate(result, now);
    carouselTick(nowMs, result.state === 'IDLE');
  }, TICK_INTERVAL_MS);

  // The one and only safety-net poll for the page's entire lifetime.
  setInterval(safeRefresh, POLL_INTERVAL_MS);

  // The one and only SSE connection for the page's entire lifetime.
  connectSSE(() => {
    safeRefresh();
  });
}

init();
