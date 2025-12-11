(function () {
  const config = window.TzSuggestionFeedbackConfig || null
  if (!config) {
    return
  }

  const tableBody = document.querySelector('#tz-suggestion-feedback-table tbody')
  const statusFilter = document.getElementById('tz-suggestion-status')
  const searchInput = document.getElementById('tz-suggestion-search')
  const refreshBtn = document.getElementById('tz-suggestion-refresh')
  const prevBtn = document.getElementById('tz-suggestion-prev')
  const nextBtn = document.getElementById('tz-suggestion-next')
  const pageInfo = document.getElementById('tz-suggestion-page-info')
  const noticeEl = document.getElementById('tz-suggestion-feedback-notice')

  const state = {
    page: 1,
    perPage: 20,
    status: '',
    search: '',
    totalPages: 1,
  }

  function showNotice(message, type = 'info') {
    if (!noticeEl) return
    noticeEl.textContent = message
    noticeEl.className = `notice notice-${type}`
    noticeEl.style.display = 'block'
    setTimeout(() => {
      noticeEl.style.display = 'none'
    }, 4000)
  }

  function buildQuery() {
    const params = new URLSearchParams()
    params.append('page', state.page)
    params.append('per_page', state.perPage)
    if (state.status) params.append('status', state.status)
    if (state.search) params.append('search', state.search)
    return params.toString()
  }

  async function fetchList() {
    if (!tableBody) return
    tableBody.innerHTML = `<tr><td colspan="7">${window.tzAdminLoading || '加载中...'}</td></tr>`

    try {
      const response = await window.fetch(`${config.listUrl}?${buildQuery()}`, {
        headers: {
          'X-WP-Nonce': config.nonce,
        },
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.message || '加载失败')
      }

      const data = await response.json()
      renderTable(data.data || [])
      updatePagination(data.pagination)
    } catch (error) {
      tableBody.innerHTML = `<tr><td colspan="7" class="has-error">${error.message}</td></tr>`
      showNotice(error.message, 'error')
    }
  }

  function renderTable(items) {
    if (!items.length) {
      tableBody.innerHTML = '<tr><td colspan="7">暂无数据</td></tr>'
      return
    }

    const rows = items
      .map(item => {
        const attachments = Array.isArray(item.attachments) && item.attachments.length
          ? item.attachments.map(file => `<li><a href="${file.url}" target="_blank" rel="noopener">${file.name || file.url}</a></li>`).join('')
          : '<li>—</li>'

        return `
          <tr data-id="${item.id}">
            <td>#${item.id}</td>
            <td>
              <div><strong>${item.full_name || '—'}</strong></div>
              <div>${item.email || ''}</div>
              <div>${item.country || ''}</div>
            </td>
            <td>
              <div>${item.product_category || '—'} / ${item.request_type || '—'}</div>
              <div>${item.order_number || ''}</div>
            </td>
            <td>
              <div class="tz-suggestion-message">${item.message || ''}</div>
              <details>
                <summary>附件 (${Array.isArray(item.attachments) ? item.attachments.length : 0})</summary>
                <ul>${attachments}</ul>
              </details>
            </td>
            <td>
              <div>${item.created_at || ''}</div>
              <div>更新：${item.updated_at || ''}</div>
            </td>
            <td>
              <span class="tz-status-badge tz-status-${item.status}">${config.labels[item.status] || item.status}</span>
            </td>
            <td>
              ${renderActions(item)}
            </td>
          </tr>
        `
      })
      .join('')

    tableBody.innerHTML = rows
    bindActionButtons()
  }

  function renderActions(item) {
    const targets = ['new', 'in_review', 'resolved', 'archived'].filter(status => status !== item.status)
    if (!targets.length) return '—'

    return `
      <div class="tz-suggestion-actions">
        ${targets
          .map(
            status => `
              <button type="button" class="button tz-suggestion-action" data-id="${item.id}" data-status="${status}">
                ${config.labels[status] || status}
              </button>
            `
          )
          .join('')}
      </div>
    `
  }

  function bindActionButtons() {
    document.querySelectorAll('.tz-suggestion-action').forEach(button => {
      button.addEventListener('click', () => {
        const id = button.dataset.id
        const status = button.dataset.status
        if (!id || !status) return
        updateStatus(id, status)
      })
    })
  }

  async function updateStatus(id, status) {
    try {
      const response = await window.fetch(`${config.listUrl}/${id}/status`, {
        method: 'PATCH',
        headers: {
          'Content-Type': 'application/json',
          'X-WP-Nonce': config.nonce,
        },
        body: JSON.stringify({ status }),
      })

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}))
        throw new Error(errorData.message || '更新失败')
      }

      showNotice('状态已更新', 'success')
      fetchList()
    } catch (error) {
      showNotice(error.message, 'error')
    }
  }

  function updatePagination(meta) {
    if (!meta) return
    state.totalPages = meta.total_pages || 1
    state.page = meta.page || 1
    if (pageInfo) {
      pageInfo.textContent = `${state.page} / ${state.totalPages}`
    }
    if (prevBtn) prevBtn.disabled = state.page <= 1
    if (nextBtn) nextBtn.disabled = state.page >= state.totalPages
  }

  function bindFilters() {
    if (refreshBtn) {
      refreshBtn.addEventListener('click', () => {
        state.status = statusFilter ? statusFilter.value : ''
        state.search = searchInput ? searchInput.value.trim() : ''
        state.page = 1
        fetchList()
      })
    }

    if (prevBtn) {
      prevBtn.addEventListener('click', () => {
        if (state.page <= 1) return
        state.page -= 1
        fetchList()
      })
    }

    if (nextBtn) {
      nextBtn.addEventListener('click', () => {
        if (state.page >= state.totalPages) return
        state.page += 1
        fetchList()
      })
    }
  }

  bindFilters()
  fetchList()
})()
