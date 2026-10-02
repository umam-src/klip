(() => {
  const target = document.querySelector('#hasil-list');
  const refresh = document.querySelector('#refresh-hasil');
  const workspace = document.querySelector('#workspace');
  const label = document.querySelector('#workspace-label');
  if (!target || !refresh || !workspace || !label) return;

  const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  const empty = (message) => `<div class="empty"><strong>${escapeHTML(message)}</strong></div>`;

  async function request(path) {
    const response = await fetch(path, { headers: { Accept: 'application/json' } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.json();
  }

  async function loadResults() {
    if (workspace.hidden) return;
    refresh.disabled = true;
    target.innerHTML = empty('Memuat hasil...');
    try {
      const spaces = await request('/api/v1/ruang');
      const name = label.textContent.replace(/^Ruang ·\s*/, '').trim();
      const ruang = spaces.find((space) => space.name === name);
      if (!ruang) throw new Error('ruang tidak ditemukan');
      const jobs = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/pekerjaan`);
      const results = [];
      await Promise.all(jobs.map(async (job) => {
        const items = await request(`/api/v1/pekerjaan/${encodeURIComponent(job.id)}/hasil`);
        for (const hasil of Array.isArray(items) ? items : []) results.push({ hasil, pekerjaan: job });
      }));
      results.sort((a, b) => String(b.hasil.created_at || '').localeCompare(String(a.hasil.created_at || '')));
      target.innerHTML = results.length ? results.map(renderResult).join('') : empty('Belum ada hasil.');
    } catch (_) {
      target.innerHTML = empty('Hasil belum dapat dimuat.');
    } finally {
      refresh.disabled = false;
    }
  }

  function renderResult(item) {
    const { hasil, pekerjaan } = item;
    const created = hasil.created_at ? new Date(hasil.created_at).toLocaleString('id-ID') : '';
    const task = hasil.tugas_id ? `Tugas · ${hasil.tugas_id}` : 'Tanpa tugas';
    const fileURL = `/api/v1/hasil/${encodeURIComponent(hasil.id)}/file`;
    const file = `<a class="button small" href="${fileURL}" target="_blank" rel="noopener">Buka hasil</a>`;
    return `<article class="item"><div class="approval-head"><strong>${escapeHTML(hasil.name)}</strong><time datetime="${escapeHTML(hasil.created_at || '')}">${escapeHTML(created)}</time></div><span>${escapeHTML(hasil.kind)} · ${escapeHTML(task)}</span><span>${escapeHTML(pekerjaan.title || pekerjaan.id)} · ${escapeHTML(hasil.path)}</span><div>${file}</div></article>`;
  }

  refresh.addEventListener('click', loadResults);
  const observer = new MutationObserver(() => {
    if (!workspace.hidden) loadResults();
  });
  observer.observe(workspace, { attributes: true, attributeFilter: ['hidden'] });
  loadResults();
})();
