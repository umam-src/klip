(() => {
  const $ = (selector) => document.querySelector(selector);

  const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);

  async function request(path, options = {}) {
    const response = await fetch(path, { ...options, headers: { Accept: 'application/json', ...(options.headers || {}) } });
    if (!response.ok) throw new Error(`request ${response.status}`);
    return response.status === 204 ? null : response.json();
  }

  async function workspaceData() {
    const spaces = await request('/api/v1/ruang');
    const name = $('#workspace-title')?.textContent?.trim();
    const ruang = spaces.find((space) => space.name === name);
    if (!ruang) throw new Error('ruang tidak ditemukan');
    const agents = await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/agen`);
    return { ruang, agents };
  }

  function fillParents(agents) {
    const select = $('#agent-parent');
    if (!select) return;
    select.innerHTML = '<option value="">Tidak ada atasan</option>';
    for (const agent of agents) {
      const option = document.createElement('option');
      option.value = agent.id;
      option.textContent = agent.name || 'Tanpa nama';
      select.appendChild(option);
    }
  }

  async function fillModels(selected = '') {
    const select = $('#agent-model');
    if (!select) return;
    select.disabled = true;
    select.innerHTML = '<option value="">Memuat model...</option>';
    try {
      const data = await request('/api/v1/provider/models');
      const models = Array.isArray(data.models) ? data.models.map((model) => String(model.id || '').trim()).filter(Boolean) : [];
      select.innerHTML = '<option value="">Pilih model</option>';
      if (selected && !models.includes(selected)) {
        const current = document.createElement('option');
        current.value = selected;
        current.textContent = selected;
        current.selected = true;
        select.appendChild(current);
      }
      for (const model of models) {
        const option = document.createElement('option');
        option.value = model;
        option.textContent = model;
        option.selected = model === selected;
        select.appendChild(option);
      }
      select.disabled = false;
    } catch (_) {
      select.innerHTML = '<option value="">Daftar model belum tersedia</option>';
    }
  }

  async function prepareForm() {
    try {
      const { agents } = await workspaceData();
      fillParents(agents);
      await fillModels('');
    } catch (_) {
      // Form tetap dapat digunakan untuk data identitas bila layanan pilihan belum tersedia.
    }
  }

  document.addEventListener('click', (event) => {
    const toggle = event.target.closest('[data-toggle="agent-form"]');
    if (!toggle) return;
    const form = $('#agent-form');
    if (!form) return;
    if (!form.hidden) return;
    prepareForm();
  }, true);

  document.addEventListener('submit', async (event) => {
    if (event.target.id !== 'agent-form') return;
    event.preventDefault();
    event.stopImmediatePropagation();
    const formElement = event.target;
    const form = new FormData(formElement);
    const submit = formElement.querySelector('button[type="submit"]');
    submit.disabled = true;
    try {
      const { ruang } = await workspaceData();
      await request(`/api/v1/ruang/${encodeURIComponent(ruang.id)}/agen`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          id: globalThis.crypto?.randomUUID ? crypto.randomUUID() : `klip-${Date.now()}`,
          name: String(form.get('name') || '').trim(),
          role: String(form.get('role') || '').trim(),
          description: String(form.get('description') || '').trim(),
          parent_id: String(form.get('parent_id') || '').trim(),
          provider_id: String(form.get('provider_id') || '').trim(),
          model_id: String(form.get('model_id') || '').trim(),
        }),
      });
      formElement.reset();
      formElement.hidden = true;
      document.querySelector('#agent-list')?.replaceChildren();
      const refresh = document.querySelector('#refresh');
      refresh?.click();
      window.setTimeout(() => document.querySelector('#workspace')?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 0);
    } catch (error) {
      alert(error?.message || 'Agen belum dapat disimpan.');
    } finally {
      submit.disabled = false;
    }
  }, true);
})();
