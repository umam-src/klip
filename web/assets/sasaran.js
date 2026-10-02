(() => {
  const $ = (selector) => document.querySelector(selector);
  const id = () => globalThis.crypto?.randomUUID?.() || `klip-${Date.now()}-${Math.random().toString(16).slice(2)}`;
  const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  const state = { ruang: null, sasaran: [] };

  async function request(path, options = {}) {
    const response = await fetch(path, { ...options, headers: { Accept: 'application/json', ...(options.headers || {}) } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.json();
  }

  async function load(ruangID) {
    state.ruang = ruangID;
    try {
      state.sasaran = await request(`/api/v1/ruang/${encodeURIComponent(ruangID)}/sasaran`);
      $('#goal-list').innerHTML = state.sasaran.length
        ? state.sasaran.map((goal) => `<article class="item"><strong>${escapeHTML(goal.title)}</strong><span>Status · ${escapeHTML(goal.status)}</span></article>`).join('')
        : '<div class="empty"><strong>Belum ada sasaran.</strong></div>';
      const select = $('#job-form select[name="sasaran_id"]');
      select.innerHTML = `<option value="">Tanpa sasaran</option>${state.sasaran.map((goal) => `<option value="${escapeHTML(goal.id)}">${escapeHTML(goal.title)}</option>`).join('')}`;
    } catch (_) {
      $('#goal-list').innerHTML = '<div class="empty"><strong>Sasaran belum dapat dimuat.</strong></div>';
    }
  }

  const observer = new MutationObserver(() => {
    const workspace = $('#workspace');
    if (workspace && !workspace.hidden && workspace.dataset.sasaranReady !== '1') {
      workspace.dataset.sasaranReady = '1';
      const label = $('#workspace-label')?.textContent || '';
      const spaces = label.split('·');
      if (window.klipRuangID) load(window.klipRuangID);
      else if (spaces.length > 1) {
        fetch('/api/v1/ruang').then((response) => response.json()).then((items) => {
          const space = items.find((item) => item.name === spaces.slice(1).join('·').trim());
          if (space) load(space.id);
        }).catch(() => {});
      }
    }
  });

  observer.observe(document.body, { subtree: true, attributes: true, attributeFilter: ['hidden'] });

  $('#goal-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    if (!state.ruang) return;
    const form = new FormData(event.currentTarget);
    try {
      await request(`/api/v1/ruang/${encodeURIComponent(state.ruang)}/sasaran`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id(), title: form.get('title') })
      });
      event.currentTarget.reset();
      event.currentTarget.hidden = true;
      await load(state.ruang);
    } catch (_) { alert('Sasaran belum dapat disimpan.'); }
  });

  $('#job-form')?.addEventListener('submit', async (event) => {
    event.preventDefault();
    event.stopImmediatePropagation();
    if (!state.ruang) return;
    const form = new FormData(event.currentTarget);
    const sasaranID = String(form.get('sasaran_id') || '').trim();
    try {
      await request(`/api/v1/ruang/${encodeURIComponent(state.ruang)}/pekerjaan`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: id(), title: form.get('title'), ...(sasaranID ? { sasaran_id: sasaranID } : {}) })
      });
      event.currentTarget.reset();
      event.currentTarget.hidden = true;
      const refreshButton = $('#refresh');
      if (refreshButton) refreshButton.click();
    } catch (_) { alert('Pekerjaan belum dapat disimpan.'); }
  }, true);
})();
