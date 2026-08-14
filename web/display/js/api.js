import { setAnchor } from './clock.js';
import { setSlides } from './carousel.js';
import { updateStaticFields } from './render.js';

let cachedDisplayData = null;

export function getCachedDisplayData() {
  return cachedDisplayData;
}

export async function refreshDisplayData() {
  const res = await fetch('/api/v1/display-data');
  if (!res.ok) throw new Error(`display-data fetch failed: ${res.status}`);
  const data = await res.json();
  setAnchor(data.now);
  cachedDisplayData = data;
  setSlides(data.slides);
  updateStaticFields(data);
  return data;
}

// Created exactly once for the page's entire lifetime by main.js — never
// recreated. The browser's native EventSource already auto-reconnects on
// a dropped connection, so no manual backoff/retry logic is needed here.
export function connectSSE(onMessage) {
  const es = new EventSource('/api/v1/sse');
  es.onmessage = onMessage;
  return es;
}
