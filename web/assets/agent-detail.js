(() => {
  const $ = (selector) => document.querySelector(selector);

  const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, (char) => ({
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  })[char]);

  const detail = (label, value) => {
    const text = value === undefined || value === null || String(value).trim() === '' ? 'Belum diatur' : value;
    return `<div><dt>${escapeHTML(label)}</dt><dd>${escapeHTML(text)}</dd></div>`;
  };

  const statusLabel = (status) => status === 'inactive' ? 'Nonaktif' : 'Aktif';

  async function request(path) {
    const response = await fetch(path, { headers: { Accept: 'application/json' } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.json();
  }

  async function openAgentDetail(button) {
    const ruangName = $('#workspace-title')?.textContent?.trim();
    const agentID = button.dataset.agent;
    if (!ruangName || !agentID) return;

    const spaces = await request('/api/v1/ruang');
    const ruang = spaces.find((space) => space.name === ruangName);
    if (!ruang) throw new Error('ruang tidak ditemukan');

    const agents = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/agen`);
    const agent = agents.find((item) => item.id === agentID);
    if (!agent) throw new Error('agen tidak ditemukan');

    const parent = agents.find((item) => item.id === agent.parent_id);
    const statusClass = agent.status === 'inactive' ? 'inactive' : '';
    const description = String(agent.description || '').trim() || 'Belum ada deskripsi agen.';

    $('#agent-title').textContent = agent.name || 'Agen';
    $('#agent-detail-card').innerHTML = `
      <div class="panel-head">
        <div>
          <h3>${escapeHTML(agent.name || 'Tanpa nama')}</h3>
          <p>${escapeHTML(agent.role || 'Agen')}</p>
        </div>
        <span class="agent-status ${statusClass}">${escapeHTML(statusLabel(agent.status))}</span>
      </div>
      <div class="stack">
        <section class="item">
          <strong>Ringkasan</strong>
          <p>${escapeHTML(description)}</p>
        </section>
        <section class="item">
          <strong>AI yang digunakan</strong>
          <dl>${detail('Penyedia', agent.provider_id)}${detail('Model', agent.model_id)}</dl>
        </section>
        <section class="item">
          <strong>Struktur agen</strong>
          <dl>${detail('Atasan', parent?.name || agent.parent_id)}${detail('ID', agent.id)}</dl>
        </section>
        <section class="item">
          <strong>Pekerjaan</strong>
          <p>Belum ada pekerjaan yang terhubung langsung ke agen.</p>
        </section>
        <section class="item">
          <strong>Aktivitas</strong>
          <p>Belum ada aktivitas agen yang ditampilkan.</p>
        </section>
        <section class="item">
          <strong>Pengaturan</strong>
          <p>Pengaturan agen akan ditambahkan pada tahap berikutnya.</p>
        </section>
      </div>`;

    $('#workspace').hidden = true;
    $('#agent-detail').hidden = false;
    $('#job-detail').hidden = true;
    $('#task-detail').hidden = true;
    $('#agent-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });
  }

  document.addEventListener('click', (event) => {
    const button = event.target.closest('[data-agent]');
    if (!button) return;

    event.preventDefault();
    event.stopImmediatePropagation();

    openAgentDetail(button).catch(() => {
      // Jika detail gagal dimuat, halaman utama tetap dapat digunakan.
    });
  }, true);
})();
