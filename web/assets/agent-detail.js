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

  async function request(path, options = {}) {
    const response = await fetch(path, { ...options, headers: { Accept: 'application/json', ...(options.headers || {}) } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.status === 204 ? null : response.json();
  }

  async function loadAgent(agenID) {
    const spaces = await request('/api/v1/ruang');
    const workspaceName = $('#workspace-title')?.textContent?.trim();
    const ruang = spaces.find((space) => space.name === workspaceName);
    if (!ruang) throw new Error('ruang tidak ditemukan');
    const agents = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/agen`);
    const agent = agents.find((item) => item.id === agenID);
    if (!agent) throw new Error('agen tidak ditemukan');
    return { ruang, agents, agent };
  }

  async function loadModels() {
    try {
      const data = await request('/api/v1/provider/models');
      return Array.isArray(data.models) ? data.models.map((model) => String(model.id || '').trim()).filter(Boolean) : [];
    } catch (_) {
      return [];
    }
  }

  function parentOptions(agents, currentID, selectedID) {
    const options = ['<option value="">Tidak ada atasan</option>'];
    for (const agent of agents) {
      if (agent.id === currentID) continue;
      const selected = agent.id === selectedID ? ' selected' : '';
      options.push(`<option value="${escapeHTML(agent.id)}"${selected}>${escapeHTML(agent.name || 'Tanpa nama')}</option>`);
    }
    return options.join('');
  }

  function modelOptions(models, selected) {
    const options = ['<option value="">Pilih model</option>'];
    if (selected && !models.includes(selected)) options.push(`<option value="${escapeHTML(selected)}" selected>${escapeHTML(selected)}</option>`);
    for (const model of models) {
      const isSelected = model === selected ? ' selected' : '';
      options.push(`<option value="${escapeHTML(model)}"${isSelected}>${escapeHTML(model)}</option>`);
    }
    return options.join('');
  }

  function renderEditForm(ruang, agents, agent) {
    return `<form id="agent-edit-form" class="inline-form agent-edit-form">
      <label>Nama<input name="name" value="${escapeHTML(agent.name)}" required maxlength="120" autocomplete="off"></label>
      <label>Peran<input name="role" value="${escapeHTML(agent.role)}" required maxlength="120" autocomplete="off"></label>
      <label>Deskripsi<textarea name="description" rows="4" maxlength="1000">${escapeHTML(agent.description || '')}</textarea></label>
      <label>Atasan<select name="parent_id">${parentOptions(agents, agent.id, agent.parent_id || '')}</select></label>
      <label>Penyedia AI<select name="provider_id"><option value="ollama"${agent.provider_id === 'ollama' ? ' selected' : ''}>Ollama</option><option value="openai-compatible"${agent.provider_id === 'openai-compatible' ? ' selected' : ''}>OpenAI-compatible</option></select></label>
      <label>Model<select name="model_id"><option value="">Memuat model...</option></select></label>
      <div class="form-actions"><button class="button primary small" type="submit">Simpan perubahan</button><button class="button small" type="button" id="cancel-agent-edit">Batal</button></div>
      <p class="form-help">Kredensial tidak diatur di Agen.</p>
    </form>`;
  }

  async function renderAgentDetail(ruang, agents, agent) {
    const parent = agents.find((item) => item.id === agent.parent_id);
    const statusClass = agent.status === 'inactive' ? 'inactive' : '';
    const description = String(agent.description || '').trim() || 'Belum ada deskripsi agen.';

    $('#agent-title').textContent = agent.name || 'Agen';
    $('#agent-detail-card').innerHTML = `
      <div class="panel-head agent-detail-head">
        <div><h3>${escapeHTML(agent.name || 'Tanpa nama')}</h3><p>${escapeHTML(agent.role || 'Agen')}</p></div>
        <div class="form-actions"><span class="agent-status ${statusClass}">${escapeHTML(statusLabel(agent.status))}</span><button class="button small" type="button" id="edit-agent">Edit</button></div>
      </div>
      <div class="stack">
        <section class="item"><strong>Ringkasan</strong><p>${escapeHTML(description)}</p></section>
        <section class="item"><strong>AI yang digunakan</strong><dl>${detail('Penyedia', agent.provider_id)}${detail('Model', agent.model_id)}</dl></section>
        <section class="item"><strong>Struktur agen</strong><dl>${detail('Atasan', parent?.name || agent.parent_id)}${detail('ID', agent.id)}</dl></section>
        <section class="item"><strong>Pekerjaan</strong><p>Belum ada pekerjaan yang terhubung langsung ke agen.</p></section>
        <section class="item"><strong>Aktivitas</strong><p>Belum ada aktivitas agen yang ditampilkan.</p></section>
        <section class="item"><strong>Pengaturan</strong><p>Gunakan Edit untuk mengubah identitas, hubungan kerja, dan AI Agen.</p></section>
      </div>`;

    $('#workspace').hidden = true;
    $('#agent-detail').hidden = false;
    $('#job-detail').hidden = true;
    $('#task-detail').hidden = true;
    $('#agent-detail').scrollIntoView({ behavior: 'smooth', block: 'start' });

    $('#edit-agent').addEventListener('click', async () => {
      $('#agent-detail-card').innerHTML = renderEditForm(ruang, agents, agent);
      const models = await loadModels();
      const modelSelect = $('#agent-edit-form [name="model_id"]');
      if (modelSelect) modelSelect.innerHTML = modelOptions(models, agent.model_id || '');
      $('#agent-edit-form [name="name"]').focus();
      $('#cancel-agent-edit').addEventListener('click', () => renderAgentDetail(ruang, agents, agent));
      $('#agent-edit-form').addEventListener('submit', async (event) => {
        event.preventDefault();
        const form = new FormData(event.currentTarget);
        const submit = event.currentTarget.querySelector('button[type="submit"]');
        submit.disabled = true;
        try {
          const updated = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/agen?agent_id=${encodeURIComponent(agent.id)}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              name: String(form.get('name') || '').trim(),
              role: String(form.get('role') || '').trim(),
              description: String(form.get('description') || '').trim(),
              parent_id: String(form.get('parent_id') || '').trim(),
              provider_id: String(form.get('provider_id') || '').trim(),
              model_id: String(form.get('model_id') || '').trim(),
            }),
          });
          const next = agents.map((item) => item.id === updated.id ? updated : item);
          await renderAgentDetail(ruang, next, updated);
        } catch (_) {
          alert('Perubahan Agen belum dapat disimpan.');
          submit.disabled = false;
        }
      });
    });
  }

  async function openAgentDetail(agenID) {
    try {
      const { ruang, agents, agent } = await loadAgent(agenID);
      await renderAgentDetail(ruang, agents, agent);
    } catch (_) {
      // Halaman utama tetap dapat digunakan jika detail gagal dimuat.
    }
  }

  document.addEventListener('click', (event) => {
    const button = event.target.closest('[data-agent]');
    if (!button) return;
    event.preventDefault();
    event.stopImmediatePropagation();
    openAgentDetail(button.dataset.agent);
  }, true);
})();
