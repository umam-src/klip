(() => {
  const state = { ruang: null, agen: null, pekerjaan: null, tugas: null, komentarParentID: null };
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

  function detail(label, value) {
    const text = value === undefined || value === null || value === '' ? 'Belum diatur' : value;
    return `<div><dt>${escapeHTML(label)}</dt><dd>${escapeHTML(text)}</dd></div>`;
  }

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
      state.agen = null; state.pekerjaan = null; state.tugas = null; state.komentarParentID = null;
      $('#workspace-label').textContent = `Ruang · ${state.ruang.name}`;
      $('#workspace-title').textContent = state.ruang.name;
      $('#workspace').hidden = false; $('#agent-detail').hidden = true; $('#job-detail').hidden = true; $('#task-detail').hidden = true;
      await Promise.all([loadAgents(), loadJobs()]);
      $('#workspace').scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (_) { list.insertAdjacentHTML('beforeend', empty('Ruang tidak dapat dibuka.')); }
  }

  async function loadAgents() {
    const target = $('#agent-list');
    try {
      const agents = await request(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/agen`);
      target.innerHTML = agents.length ? agents.map((agent) => `<button class="item item-button" type="button" data-agent="${escapeHTML(agent.id)}"><strong>${escapeHTML(agent.name)}</strong><span>${escapeHTML(agent.provider_id || 'Penyedia belum diatur')}${agent.model_id ? ` · ${escapeHTML(agent.model_id)}` : ''}</span></button>`).join('') : empty('Belum ada agen.');
      target.querySelectorAll('[data-agent]').forEach((button) => button.addEventListener('click', () => openAgent(button.dataset.agent, agents)));
    } catch (_) { target.innerHTML = empty('Agen belum dapat dimuat.'); }
  }

  function openAgent(agenID, agents) {
    state.agen = agents.find((agent) => agent.id === agenID);
    if (!state.agen) return;
    $('#agent-title').textContent = state.agen.name;
    $('#agent-detail-card').innerHTML = `<dl>${detail('Nama', state.agen.name)}${detail('Deskripsi', state.agen.description)}${detail('Penyedia', state.agen.provider_id)}${detail('Model', state.agen.model_id)}${detail('ID', state.agen.id)}</dl>`;
    $('#workspace').hidden = true; $('#agent-detail').hidden = false; $('#job-detail').hidden = true; $('#task-detail').hidden = true;
    $('#agent-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  async function loadJobs() {
    const target = $('#job-list');
    try {
      const jobs = await request(`/api/v1/ruang/${encodeURIComponent(state.ruang.id)}/pekerjaan`);
      target.innerHTML = jobs.length ? jobs.map((job) => `<button class="item item-button" type="button" data-job="${escapeHTML(job.id)}"><strong>${escapeHTML(job.title)}</strong><span>Status · ${escapeHTML(job.status)}</span></button>`).join('') : empty('Belum ada pekerjaan.');
      target.querySelectorAll('[data-job]').forEach((button) => button.addEventListener('click', () => openJob(button.dataset.job, jobs)));
    } catch (_) { target.innerHTML = empty('Pekerjaan belum dapat dimuat.'); }
  }

  async function openJob(pekerjaanID, jobs) {
    state.pekerjaan = jobs.find((job) => job.id === pekerjaanID);
    if (!state.pekerjaan) return;
    state.tugas = null; state.komentarParentID = null;
    try {
      $('#job-title').textContent = state.pekerjaan.title;
      $('#workspace').hidden = true; $('#agent-detail').hidden = true; $('#job-detail').hidden = false; $('#task-detail').hidden = true;
      await loadTasks();
      $('#job-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });
    } catch (_) { $('#task-list').innerHTML = empty('Pekerjaan tidak dapat dibuka.'); }
  }

  async function loadTasks() {
    const target = $('#task-list');
    try {
      const tasks = await request(`/api/v1/pekerjaan/${encodeURIComponent(state.pekerjaan.id)}/tugas`);
      target.innerHTML = tasks.length ? tasks.map((task) => `<button class="item item-button" type="button" data-task="${escapeHTML(task.id)}"><strong>${escapeHTML(task.title)}</strong><span>Status · ${escapeHTML(task.status)}</span></button>`).join('') : empty('Belum ada tugas.');
      target.querySelectorAll('[data-task]').forEach((button) => button.addEventListener('click', () => openTask(button.dataset.task, tasks)));
    } catch (_) { target.innerHTML = empty('Tugas belum dapat dimuat.'); }
  }

  async function openTask(tugasID, tasks) {
    state.tugas = tasks.find((task) => task.id === tugasID);
    if (!state.tugas) return;
    state.komentarParentID = null;
    $('#task-title').textContent = state.tugas.title;
    $('#task-detail-card').innerHTML = `<dl>${detail('Judul', state.tugas.title)}${detail('Status', state.tugas.status)}${detail('Posisi', state.tugas.position)}${detail('Tugas induk', state.tugas.parent_id)}${detail('ID', state.tugas.id)}</dl>`;
    clearReply();
    $('#workspace').hidden = true; $('#agent-detail').hidden = true; $('#job-detail').hidden = true; $('#task-detail').hidden = false;
    await loadComments();
    $('#task-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  function commentDepth(comment, commentsByID) {
    let depth = 0;
    const seen = new Set();
    let parentID = comment.parent_id;
    while (parentID && !seen.has(parentID)) {
      seen.add(parentID);
      const parent = commentsByID.get(parentID);
      if (!parent) break;
      depth += 1;
      parentID = parent.parent_id;
    }
    return Math.min(depth, 4);
  }

  function renderComments(comments) {
    const target = $('#comment-list');
    if (!Array.isArray(comments) || !comments.length) {
      target.innerHTML = empty('Belum ada komentar.');
      return;
    }
    const commentsByID = new Map(comments.map((comment) => [comment.id, comment]));
    target.innerHTML = comments.map((comment) => {
      const depth = commentDepth(comment, commentsByID);
      const parent = comment.parent_id ? commentsByID.get(comment.parent_id) : null;
      const context = parent ? `Membalas: ${escapeHTML(parent.body.slice(0, 100))}` : '';
      const created = comment.created_at ? new Date(comment.created_at).toLocaleString('id-ID') : '';
      return `<article class="comment" style="--comment-depth:${depth}">
        <div class="comment-meta"><strong>Komentar</strong><time datetime="${escapeHTML(comment.created_at || '')}">${escapeHTML(created)}</time></div>
        ${context ? `<div class="comment-context">${context}</div>` : ''}
        <p>${escapeHTML(comment.body)}</p>
        <button class="button small" type="button" data-reply="${escapeHTML(comment.id)}">Balas</button>
      </article>`;
    }).join('');
    target.querySelectorAll('[data-reply]').forEach((button) => button.addEventListener('click', () => beginReply(button.dataset.reply, commentsByID)));
  }

  async function loadComments() {
    const target = $('#comment-list');
    target.innerHTML = empty('Memuat komentar...');
    try {
      const comments = await request(`/api/v1/tugas/${encodeURIComponent(state.tugas.id)}/komentar`);
      renderComments(comments);
    } catch (_) {
      target.innerHTML = empty('Komentar belum dapat dimuat.');
    }
  }

  function beginReply(commentID, commentsByID) {
    const parent = commentsByID.get(commentID);
    if (!parent) return;
    state.komentarParentID = parent.id;
    $('#replying-to').textContent = `Membalas: ${parent.body.slice(0, 120)}`;
    $('#replying-to').hidden = false;
    $('#cancel-reply').hidden = false;
    $('#comment-form textarea').focus();
  }

  function clearReply() {
    state.komentarParentID = null;
    const context = $('#replying-to');
    if (context) { context.textContent = ''; context.hidden = true; }
    $('#cancel-reply').hidden = true;
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
    $('#comment-form').addEventListener('submit', async (event) => {
      event.preventDefault();
      const form = new FormData(event.currentTarget);
      const body = String(form.get('body') || '').trim();
      if (!body || !state.tugas) return;
      const submit = event.currentTarget.querySelector('button[type="submit"]');
      submit.disabled = true;
      try {
        await send(`/api/v1/tugas/${encodeURIComponent(state.tugas.id)}/komentar`, { id: id(), parent_id: state.komentarParentID || undefined, body });
        event.currentTarget.reset();
        clearReply();
        await loadComments();
      } catch (_) { alert('Komentar belum dapat disimpan.'); }
      finally { submit.disabled = false; }
    });
  }

  $('#back-to-spaces').addEventListener('click', () => { $('#workspace').hidden = true; $('#agent-detail').hidden = true; $('#job-detail').hidden = true; $('#task-detail').hidden = true; $('#ruang').scrollIntoView({ behavior: 'smooth' }); });
  $('#back-to-workspace-from-agent').addEventListener('click', () => { $('#agent-detail').hidden = true; $('#workspace').hidden = false; $('#workspace').scrollIntoView({ behavior: 'smooth', block: 'start' }); });
  $('#back-to-workspace').addEventListener('click', () => { $('#job-detail').hidden = true; $('#workspace').hidden = false; $('#workspace').scrollIntoView({ behavior: 'smooth' }); });
  $('#back-to-job').addEventListener('click', () => { $('#task-detail').hidden = true; $('#job-detail').hidden = false; $('#job-detail').scrollIntoView({ behavior: 'smooth', block: 'start' }); });
  $('#refresh-comments').addEventListener('click', loadComments);
  $('#cancel-reply').addEventListener('click', clearReply);
  refresh.addEventListener('click', loadSpaces);
  bindForms();
  loadSpaces();
})();
