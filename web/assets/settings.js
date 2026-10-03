(() => {
  const $ = (selector) => document.querySelector(selector);

  function escapeHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  async function requestJSON(url, options = {}) {
    const response = await fetch(url, {
      ...options,
      headers: { Accept: 'application/json', 'Content-Type': 'application/json', ...(options.headers || {}) }
    });
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(data.error || `Permintaan gagal (${response.status}).`);
      error.data = data;
      throw error;
    }
    return data;
  }

  function setMessage(text, type = '') {
    const target = $('#settings-message');
    if (!target) return;
    target.className = `settings-message ${type}`.trim();
    target.textContent = text;
    target.hidden = !text;
  }

  async function loadSettings() {
    const target = $('#settings-list');
    if (!target) return;
    target.innerHTML = '<div class="empty"><strong>Memuat pengaturan...</strong></div>';
    try {
      const settings = await requestJSON('/api/v1/settings', { headers: { 'Content-Type': 'application/json' } });
      const provider = $('#settings-provider');
      const baseURL = $('#settings-base-url');
      const model = $('#settings-model');
      if (provider) provider.value = settings.provider || 'ollama';
      if (baseURL) baseURL.value = settings.base_url || '';
      if (model && settings.model) {
        if (!Array.from(model.options).some((option) => option.value === settings.model)) {
          model.insertAdjacentHTML('beforeend', `<option value="${escapeHTML(settings.model)}">${escapeHTML(settings.model)}</option>`);
        }
        model.value = settings.model;
      }
      target.innerHTML = `<dl class="settings-grid"><div><dt>Bahasa</dt><dd>${escapeHTML(settings.locale)}</dd></div><div><dt>Penyedia AI</dt><dd>${escapeHTML(settings.provider)}</dd></div><div><dt>Model</dt><dd>${escapeHTML(settings.model || 'Belum diatur')}</dd></div></dl>`;
      setMessage('');
    } catch (_) {
      target.innerHTML = '<div class="empty"><strong>Pengaturan belum dapat dimuat.</strong></div>';
      setMessage('Pengaturan belum dapat dimuat.', 'error');
    }
  }

  async function checkProvider() {
    const provider = $('#settings-provider')?.value.trim();
    const baseURL = $('#settings-base-url')?.value.trim();
    if (!provider || !baseURL) {
      setMessage('Penyedia dan alamat wajib diisi.', 'error');
      return;
    }
    setMessage('Memeriksa koneksi...');
    try {
      const status = await requestJSON('/api/v1/provider/status', {
        method: 'POST',
        body: JSON.stringify({ provider, base_url: baseURL })
      });
      if (!status.reachable) throw new Error('Penyedia belum dapat dihubungi.');
      setMessage('Koneksi berhasil.', 'ready');
    } catch (error) {
      const belumTerhubung = error.data && error.data.reachable === false;
      setMessage(belumTerhubung ? 'Penyedia belum dapat dihubungi.' : (error.message || 'Koneksi belum tersedia.'), 'error');
    }
  }

  async function loadModels() {
    const provider = $('#settings-provider')?.value.trim();
    const baseURL = $('#settings-base-url')?.value.trim();
    const select = $('#settings-model');
    if (!provider || !baseURL || !select) {
      setMessage('Penyedia dan alamat wajib diisi.', 'error');
      return;
    }
    select.disabled = true;
    select.innerHTML = '<option value="">Memuat model...</option>';
    setMessage('Mengambil daftar model...');
    try {
      const data = await requestJSON('/api/v1/provider/models', {
        method: 'POST',
        body: JSON.stringify({ provider, base_url: baseURL })
      });
      const models = Array.isArray(data.models) ? data.models.filter((item) => String(item.id || '').trim()) : [];
      if (!models.length) {
        select.innerHTML = '<option value="">Tidak ada model</option>';
        setMessage('Penyedia tidak mengembalikan model.', 'error');
        return;
      }
      select.innerHTML = '<option value="">Pilih model</option>';
      for (const item of models) {
        const id = String(item.id).trim();
        const option = document.createElement('option');
        option.value = id;
        option.textContent = id;
        select.appendChild(option);
      }
      select.disabled = false;
      setMessage(`${models.length} model tersedia.`, 'ready');
    } catch (error) {
      select.innerHTML = '<option value="">Daftar model belum tersedia</option>';
      setMessage(error.message || 'Daftar model belum tersedia.', 'error');
    }
  }

  async function saveSettings() {
    const provider = $('#settings-provider')?.value.trim();
    const baseURL = $('#settings-base-url')?.value.trim();
    const model = $('#settings-model')?.value.trim();
    if (!provider || !baseURL || !model) {
      setMessage('Penyedia, alamat, dan model wajib dipilih.', 'error');
      return;
    }
    const button = $('#save-settings');
    if (button) button.disabled = true;
    setMessage('Menyimpan pengaturan...');
    try {
      await requestJSON('/api/v1/settings', {
        method: 'PUT',
        body: JSON.stringify({ provider, base_url: baseURL, model })
      });
      setMessage('Pengaturan tersimpan.', 'ready');
      await loadSettings();
      await loadModels();
    } catch (error) {
      setMessage(error.message || 'Pengaturan belum dapat disimpan.', 'error');
    } finally {
      if (button) button.disabled = false;
    }
  }

  document.addEventListener('DOMContentLoaded', () => {
    loadSettings();
    $('#refresh-settings')?.addEventListener('click', loadSettings);
    $('#check-provider')?.addEventListener('click', checkProvider);
    $('#load-models')?.addEventListener('click', loadModels);
    $('#save-settings')?.addEventListener('click', saveSettings);
  });
})();
