import { test } from 'node:test';
import assert from 'node:assert/strict';
import { formatCountdown } from './format.js';

test('formatCountdown: under a minute', () => {
  assert.equal(formatCountdown(5), '00:05');
});

test('formatCountdown: minutes and seconds, under an hour', () => {
  assert.equal(formatCountdown(11 * 60 + 59), '11:59');
});

test('formatCountdown: rolls over to H:MM:SS past one hour', () => {
  // 101 minutes 10 seconds -> 1 hour, 41 minutes, 10 seconds.
  assert.equal(formatCountdown(101 * 60 + 10), '1:41:10');
});

test('formatCountdown: multi-hour', () => {
  assert.equal(formatCountdown(5 * 3600 + 2 * 60 + 3), '5:02:03');
});

test('formatCountdown: negative input clamps to zero', () => {
  assert.equal(formatCountdown(-5), '00:00');
});
