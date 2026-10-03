(() => {
  const $ = (selector) => document.querySelector(selector);

  async function loadProviderStatus() {
    const target = $('#provider-status');
    if (!target) return;
    target.innerHTML = '<div class="empty"><strong>Memuat status penyedia...</strong></div>';
    try {
      const response = await fetch('/api/v1/provider/status', { headers: { Accept: 'application/json' } });
      const status = await response.json();
      if (!response.ok && response.status !== 503) throw new Error(`request ${response.status}`);
      const reachable = status.reachable ? 'Terhubung' : 'Belum terhubung';
      const configured = status.configured ? 'Sudah diatur' : 'Belum diatur';
      target.innerHTML = `<div class="provider-status-grid">
        <div><dt>Penyedia</dt><dd>${escapeHTML(status.provider || 'Belum diatur')}</dd></div>
        <div><dt>Konfigurasi model</dt><dd>${configured}</dd></div>
        <div><dt>Koneksi</dt><dd>${reachable}</dd></div>
      </div>`;
    } catch (_) {
      target.innerHTML = '<div class="empty"><strong>Status penyedia belum dapat dimuat.</strong></div>';
    }
  }

  async function loadModels() {
    const select = $('#provider-model');
    if (!select) return;
    select.disabled = true;
    select.innerHTML = '<option value="">Memuat model...</option>';
    try {
      const response = await fetch('/api/v1/provider/models', { headers: { Accept: 'application/json' } });
      const data = await response.json();
      if (!response.ok) throw new Error(data.error || `request ${response.status}`);
      const models = Array.isArray(data.models) ? data.models : [];
      if (models.length === 0) {
        select.innerHTML = '<option value="">Tidak ada model</option>';
        return;
      }
      select.innerHTML = '<option value="">Pilih model</option>';
      for (const model of models) {
        const id = String(model.id || '').trim();
        if (!id) continue;
        const option = document.createElement('option');
        option.value = id;
        option.textContent = id;
        select.appendChild(option);
      }
      select.disabled = false;
      select.dispatchEvent(new Event('change'));
    } catch (error) {
      select.innerHTML = `<option value="">${escapeHTML(error.message || 'Daftar model belum tersedia')}</option>`;
    }
  }

  function applyModelToAgent(model) {
    const input = document.querySelector('#agent-form [name="model_id"]');
    if (input) input.value = model;
  }

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  document.addEventListener('DOMContentLoaded', () => {
    loadProviderStatus();
    loadModels();
    $('#refresh-provider')?.addEventListener('click', () => {
      loadProviderStatus();
      loadModels();
    });
    $('#provider-model')?.addEventListener('change', (event) => {
      const model = event.target.value;
      if (model) applyModelToAgent(model);
    });
  });
})();
