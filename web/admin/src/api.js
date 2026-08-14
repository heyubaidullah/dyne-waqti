const BASE = '/api/v1'

export class AuthError extends Error {
  constructor() {
    super('unauthorized')
  }
}

export class RateLimitError extends Error {
  constructor(retryAfterSeconds) {
    super('rate limited')
    this.retryAfterSeconds = retryAfterSeconds
  }
}

async function request(path, opts = {}) {
  const res = await fetch(BASE + path, {
    credentials: 'include',
    ...opts,
  })

  if (res.status === 401) throw new AuthError()
  if (res.status === 429) {
    throw new RateLimitError(Number(res.headers.get('Retry-After')) || null)
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || res.statusText)
  }
  if (res.status === 204) return null
  return res.json()
}

function jsonBody(data) {
  return {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  }
}

export const api = {
  checkSession: () => request('/admin/session'),
  login: (passphrase) => request('/auth/login', jsonBody({ passphrase })),
  logout: () => request('/auth/logout', { method: 'POST' }),

  getSettings: () => request('/admin/settings'),
  updateSettings: (settings) => request('/admin/settings', jsonBody(settings)),

  updatePrayerTimes: (payload) => request('/admin/prayer-times', jsonBody(payload)),

  setBlackout: (active) => request('/admin/blackout', jsonBody({ active })),
  publishJanazah: (notice) => request('/admin/janazah', jsonBody({ action: 'publish', ...notice })),
  dismissJanazah: () => request('/admin/janazah', jsonBody({ action: 'dismiss' })),

  listSlides: () => request('/admin/slides'),
  uploadSlide: (formData) => request('/admin/slides', { method: 'POST', body: formData }),
  updateSlide: (id, payload) =>
    request(`/admin/slides/${id}`, { method: 'PATCH', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) }),
  deleteSlide: (id) => request(`/admin/slides/${id}`, { method: 'DELETE' }),

  uploadLogo: (formData) => request('/admin/logo', { method: 'POST', body: formData }),
  deleteLogo: () => request('/admin/logo', { method: 'DELETE' }),

  getDisplayData: () => request('/display-data'),
}
