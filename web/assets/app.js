(() => {
  const state = { ruang: null, pekerjaan: null };
  const $ = (selector) => document.querySelector(selector);
  const list = $('#space-list');
  const refresh = $('#refresh');

  async function request(path, options = {}) {
    const response = await fetch(path, { ...options, headers: { Accept: 'application/json', ...(options.headers || {}) } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.status === 204 ? null : response.json();
  }

  async function send(path, body) {
    return request(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });
  }

  function id() {
    if (globalThis.crypto?.randomUUID) return crypto.randomUUID();
    return `klip-${Date.now()}-${Math.random().toString(16).slice(2)}`;
  }

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  function empty(message) { return `<div class="empty"><strong>${escapeHTML(message)}</strong></div>`; }

  async function loadSpaces() {
    refresh.disabled = true;
    try {
      const spaces = await request('/api/v1/ruang');
      list.innerHTML = Array.isArray(spaces) && spaces.length
        ? spaces.map((space) => `<button class="card card-button" type="button" data-space="${escapeHTML(space.id)}"><h3>${escapeHTML(space.name || 'Tanpa nama')}</h3><p>${escapeHTML(space.id || '')}</p></button>`).join('')
        : empty('Belum ada ruang.');
      list.querySelectorAll('[data-space]').forEach((button) => button.addEventListener('click', () => openSpace(button.dataset.space)));
    } catch (_) { list.innerHTML = empty('Ruang belum dapat dimuat. Pastikan layanan Klip sedang berjalan.'); }
    finally { refresh.disabled = false; }
  }

  async function openSpace(ruangID) {
    try {
      const spaces = await request('/api/v1/ruang');
      state.ruang = spaces.find((space) => space.id === ruangID);
      if (!state.ruang) throw new Error('ruang tidak ditemukan');
      $('#workspace-label').textContent = `Ruang · ${state.ruang.name}`;
      $('#workspace-title').textContent = state.ruang.name;
      $('#workspace').hidden = false; $('#job-detail').hidden = true;
      await Promise.all([loadAgents(), loadJobs()]);
      $('#workspace').scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (_) { list.insertAdjacentHTML('beforeend', empty('Ruang tidak dapat dibuka.')); }
  }

  async function loadAgents() {
    const target = $('#agent-list');
    try {
      const agents = await request(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/agen`);
      target.innerHTML = agents.length ? agents.map((agent) => `<article class="item"><strong>${escapeHTML(agent.name)}</strong><span>${escapeHTML(agent.provider_id || 'Penyedia belum diatur')}${agent.model_id ? ` · ${escapeHTML(agent.model_id)}` : ''}</span></article>`).join('') : empty('Belum ada agen.');
    } catch (_) { target.innerHTML = empty('Agen belum dapat dimuat.'); }
  }

  async function loadJobs() {
    const target = $('#job-list');
    try {
      const jobs = await request(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/pekerjaan`);
      target.innerHTML = jobs.length ? jobs.map((job) => `<button class="item item-button" type="button" data-job="${escapeHTML(job.id)}"><strong>${escapeHTML(job.title)}</strong><span>Status · ${escapeHTML(job.status)}</span></button>`).join('') : empty('Belum ada pekerjaan.');
      target.querySelectorAll('[data-job]').forEach((button) => button.addEventListener('click', () => openJob(button.dataset.job)));
    } catch (_) { target.innerHTML = empty('Pekerjaan belum dapat dimuat.'); }
  }

  async function openJob(pekerjaanID) {
    try {
      const jobs = await request(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/pekerjaan`);
      state.pekerjaan = jobs.find((job) => job.id === pekerjaanID);
      if (!state.pekerjaan) throw new Error('pekerjaan tidak ditemukan');
      $('#job-title').textContent = state.pekerjaan.title;
      $('#job-detail').hidden = false;
      await loadTasks();
      $('#job-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (_) { $('#job-detail').hidden = false; $('#task-list').innerHTML = empty('Pekerjaan tidak dapat dibuka.'); }
  }

  async function loadTasks() {
    const target = $('#task-list');
    try {
      const tasks = await request(`/api/v1/pekerjaan/${encodeURIComponent(state.pekerjaan.id)}/tugas`);
      target.innerHTML = tasks.length ? tasks.map((task) => `<article class="item"><strong>${escapeHTML(task.title)}</strong><span>Status · ${escapeHTML(task.status)}</span></article>`).join('') : empty('Belum ada tugas.');
    } catch (_) { target.innerHTML = empty('Tugas belum dapat dimuat.'); }
  }

  function bindForms() {
    document.querySelectorAll('[data-toggle]').forEach((button) => button.addEventListener('click', () => {
      const form = $(`#${button.dataset.toggle}`);
      form.hidden = !form.hidden;
      if (!form.hidden) form.querySelector('input')?.focus();
    }));
    $('#agent-form').addEventListener('submit', async (event) => {
      event.preventDefault(); const form = new FormData(event.currentTarget);
      try { await send(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/agen`, { id: id(), name: form.get('name'), provider_id: form.get('provider_id'), model_id: form.get('model_id') }); event.currentTarget.reset(); event.currentTarget.hidden = true; await loadAgents(); }
      catch (_) { alert('Agen belum dapat disimpan.'); }
    });
    $('#job-form').addEventListener('submit', async (event) => {
      event.preventDefault(); const form = new FormData(event.currentTarget);
      try { await send(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/pekerjaan`, { id: id(), title: form.get('title') }); event.currentTarget.reset(); event.currentTarget.hidden = true; await loadJobs(); }
      catch (_) { alert('Pekerjaan belum dapat disimpan.'); }
    });
    $('#task-form').addEventListener('submit', async (event) => {
      event.preventDefault(); const form = new FormData(event.currentTarget);
      try { await send(`/api/v1/pekerjaan/${encodeURIComponent(state.pekerjaan.id)}/tugas`, { id: id(), title: form.get('title'), position: 0 }); event.currentTarget.reset(); event.currentTarget.hidden = true; await loadTasks(); }
      catch (_) { alert('Tugas belum dapat disimpan.'); }
    });
  }

  $('#back-to-spaces').addEventListener('click', () => { $('#workspace').hidden = true; $('#job-detail').hidden = true; $('#ruang').scrollIntoView({ behavior: 'smooth' }); });
  $('#back-to-workspace').addEventListener('click', () => { $('#job-detail').hidden = true; $('#workspace').scrollIntoView({ behavior: 'smooth' }); });
  refresh.addEventListener('click', loadSpaces);
  bindForms();
  loadSpaces();
})();
