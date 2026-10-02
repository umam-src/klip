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

  function escapeHTML(value) {
    return String(value).replace(/[&<>"']/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
  }

  document.addEventListener('DOMContentLoaded', loadProviderStatus);
})();
