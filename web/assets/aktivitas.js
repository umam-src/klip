(() => {
  const target = document.querySelector('#activity-list');
  const refresh = document.querySelector('#refresh-activity');
  const workspace = document.querySelector('#workspace');
  const label = document.querySelector('#workspace-label');
  if (!target || !refresh || !workspace || !label) return;

  const escapeHTML = (value) => String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  const empty = (message) => `<div class="empty"><strong>${escapeHTML(message)}</strong></div>`;

  async function request(path) {
    const response = await fetch(path, { headers: { Accept: 'application/json' } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.json();
  }

  async function loadActivity() {
    if (workspace.hidden) return;
    refresh.disabled = true;
    target.innerHTML = empty('Memuat riwayat...');
    try {
      const spaces = await request('/api/v1/ruang');
      const name = label.textContent.replace(/^Ruang ·\s*/, '').trim();
      const ruang = spaces.find((space) => space.name === name);
      if (!ruang) throw new Error('ruang tidak ditemukan');

      const jobs = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/pekerjaan`);
      const events = [];
      await Promise.all(jobs.map(async (job) => {
        const items = await request(`/api/v1/pekerjaan/${encodeURIComponent(job.id)}/aktivitas?limit=50`);
        for (const event of Array.isArray(items) ? items : []) events.push({ event, pekerjaan: job });
      }));
      events.sort((a, b) => String(b.event.created_at || '').localeCompare(String(a.event.created_at || '')));
      target.innerHTML = events.length ? events.slice(0, 50).map(renderActivity).join('') : empty('Belum ada aktivitas.');
    } catch (_) {
      target.innerHTML = empty('Riwayat aktivitas belum dapat dimuat.');
    } finally {
      refresh.disabled = false;
    }
  }

  function renderActivity(item) {
    const { event, pekerjaan } = item;
    const created = event.created_at ? new Date(event.created_at).toLocaleString('id-ID') : '';
    const summary = event.summary || event.type || 'Aktivitas';
    const task = event.tugas_id ? ` · Tugas ${event.tugas_id}` : '';
    return `<article class="activity-item item">
      <div class="activity-head"><strong>${escapeHTML(summary)}</strong><time datetime="${escapeHTML(event.created_at || '')}">${escapeHTML(created)}</time></div>
      <div class="activity-meta">${escapeHTML(pekerjaan.title || pekerjaan.id)}${escapeHTML(task)}</div>
    </article>`;
  }

  refresh.addEventListener('click', loadActivity);
  const observer = new MutationObserver(() => {
    if (!workspace.hidden) loadActivity();
  });
  observer.observe(workspace, { attributes: true, attributeFilter: ['hidden'] });
  loadActivity();
})();
